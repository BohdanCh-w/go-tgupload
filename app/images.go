package app

import (
	"io/ioutil"
	"path/filepath"
	"sync"
)

func (app *App) uploadImages(preImages map[string]string) (map[string]string, bool) {
	imageFiles, err := ioutil.ReadDir(app.Cfg.PathToImgFolder)
	if err != nil {
		app.Cfg.Logger.Fatal(err)
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
		ok = true
	)

	images := map[string]string{}

	for _, file := range imageFiles {
		path := filepath.Join(app.Cfg.PathToImgFolder, file.Name())

		if _, ok := preImages[path]; ok {
			images[path] = preImages[path]
			continue
		}

		wg.Add(1)

		go func(path string) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			images[path], err = postImage(path)
			if err != nil {
				ok = false
				delete(images, path)
				app.Cfg.Logger.Printf("PostImage failed with error: %v on path: %s", err, path)
			}
		}(path)
	}

	wg.Wait()
	return images, ok
}
