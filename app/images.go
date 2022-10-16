package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"
)

func (app *App) uploadImages(cache map[string]string) (map[string]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	imageFiles, err := ioutil.ReadDir(app.Cfg.PathToImgFolder)
	if err != nil {
		app.Cfg.Logger.Fatal(err)
	}

	var (
		sem     = semaphore.NewWeighted(5)
		images  = make(map[string]string)
		imgChan = make(chan Image)
		wg      sync.WaitGroup
		done    = make(chan struct{})
		mErr    multierror.Error
	)

	go func() { // collector
		defer close(done)

		for img := range imgChan {
			images[img.path] = img.url

			multierror.Append(&mErr, img.err)
		}
	}()

	for _, f := range imageFiles { // producer
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}

		wg.Add(1)

		file := f

		go func() {
			defer wg.Done()
			defer sem.Release(1)

			img := Image{
				path: filepath.Join(app.Cfg.PathToImgFolder, file.Name()),
			}

			defer func() {
				imgChan <- img
			}()

			if _, ok := cache[img.path]; ok {
				img.url = cache[img.path]

				return
			}

			defer delete(cache, img.path)

			data, err := os.ReadFile(img.path)
			if err != nil {
				img.err = fmt.Errorf("read image %s: %w", file.Name(), err)

				return
			}

			res, err := postImage(file.Name(), data)
			if err != nil {
				img.err = fmt.Errorf("post image %s: %w", file.Name(), err)

				return
			}

			img.url = res

			return
		}()
	}

	wg.Wait()

	close(imgChan)

	<-done

	return images, mErr.ErrorOrNil()
}

type Image struct {
	path string
	data []byte
	hash string
	url  string

	err error
}
