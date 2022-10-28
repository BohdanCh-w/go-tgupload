package post

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/browser"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	"github.com/zumorl/go-tgupload/config"
	"github.com/zumorl/go-tgupload/entities"
	"github.com/zumorl/go-tgupload/pkg/utils"
	"github.com/zumorl/go-tgupload/services"
	"github.com/zumorl/go-tgupload/usecases"
)

type poster struct {
	logger *zap.Logger
	cdn    services.CDN
	tgAPI  services.TelegraphAPI
}

func (p *poster) post(ctx context.Context, cfg config.Config) error {
	pCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	images, err := listImages(cfg.PathToImgFolder, cfg.TitleImgPath, cfg.CaptionImgPath)
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	urls, err := uploadImages(pCtx, p.cdn, images)
	if err != nil {
		return fmt.Errorf("upload images: %w", err)
	}

	page := generatePage(cfg.Title, cfg.AuthorName, cfg.AuthorURL, urls)
	pageURL, err := p.tgAPI.CreatePage(ctx, page)
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	return generateOutput(pageURL, cfg.PathToOutputFile, cfg.AutoOpen)
}

func uploadImages(ctx context.Context, cdn services.CDN, imgFiles []entities.MediaFile) ([]string, error) {
	var (
		sem          = semaphore.NewWeighted(1)
		uploaded     = make([]entities.MediaFile, 0, len(imgFiles))
		uploadedURLs = make([]string, 0, len(imgFiles))
		resChan      = make(chan uploadResult)
		wg           sync.WaitGroup
		done         = make(chan struct{})
		mErr         multierror.Error
	)

	go func() { // collector
		defer close(done)

		for res := range resChan {
			uploaded = append(uploaded, res.img)
			multierror.Append(&mErr, res.err)
		}
	}()

	for i := range imgFiles { // producer
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}

		wg.Add(1)

		idx := i

		go func() {
			defer wg.Done()
			defer sem.Release(1)

			res := uploadResult{
				img: imgFiles[idx],
			}

			defer func() {
				resChan <- res
			}()

			url, err := cdn.Upload(ctx, imgFiles[idx])
			if err != nil {
				res.err = fmt.Errorf("post image %s: %w", imgFiles[idx].Name, err)
			}

			res.img.URL = url
		}()
	}

	if mErr.ErrorOrNil() != nil {
		return nil, fmt.Errorf("upload images: %w", &mErr)
	}

	orderMediaToOriginal(uploaded, imgFiles)

	for _, img := range uploaded {
		uploadedURLs = append(uploadedURLs, img.Name)
	}

	return uploadedURLs, nil
}

// orders files according to the original order. Comparing by name
func orderMediaToOriginal(files, original []entities.MediaFile) {
	order := make(map[string]int, len(original))
	for i, val := range original {
		order[val.Name] = i
	}

	sort.Slice(files, func(i, j int) bool {
		return order[files[i].Name] < order[files[j].Name]
	})
}

type uploadResult struct {
	img entities.MediaFile
	err error
}

func listImages(dir string, titles, captions []string) ([]entities.MediaFile, error) {
	imageFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read img directory: %w", err)
	}

	pathes := make([]string, 0, len(titles)+len(imageFiles)+len(captions))
	pathes = append(pathes, titles...)

	for _, file := range imageFiles {
		if file.IsDir() {
			continue
		}

		if !usecases.IsImage(file.Name()) {
			continue
		}

		pathes = append(pathes, filepath.Join(dir, file.Name()))
	}

	pathes = append(pathes, captions...)

	images := make([]entities.MediaFile, 0, len(imageFiles))

	for _, file := range pathes {
		img, err := usecases.LoadMedia(file)
		if err != nil {
			return images, fmt.Errorf("load image: %w", err)
		}

		images = append(images, img)
	}

	return images, nil
}

func generatePage(title, authorName, authorURL string, imgURLs []string) entities.Page {
	res := entities.Page{
		Title:      title,
		AuthorName: utils.PtrOrNil(authorName),
		AuthorURL:  utils.PtrOrNil(authorURL),
		Content:    make([]entities.Node, 0, len(imgURLs)),
	}

	for _, url := range imgURLs {
		res.Content = append(res.Content, entities.Node{
			Tag:   "img",
			Attrs: map[string]string{"src": url},
		})
	}

	return res
}

func generateOutput(url, outputPath string, autoOpen bool) error {
	_, _ = os.Stdout.WriteString(fmt.Sprintf("Article posted: %s", url))

	if autoOpen {
		if err := browser.OpenURL(url); err != nil {
			return fmt.Errorf("open url: %w", err)
		}
	}

	if len(outputPath) != 0 {
		if err := os.WriteFile(outputPath, []byte(url), 0o666); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	return nil
}
