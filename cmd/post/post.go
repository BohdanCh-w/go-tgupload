package post

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	"github.com/sqweek/dialog"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/usecases"
)

type poster struct {
	uploader *usecases.CDNUploader
	tgAPI    services.TelegraphAPI
}

func (p *poster) post(ctx context.Context, dir, title string, noDialog, autoOpen bool) error {
	pCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	images, err := listImages(dir)
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	if title == "" {
		title, err = utils.RepeatPrompt(ctx, "Enter title", "")
		if err != nil {
			return fmt.Errorf("prompt title: %w", err)
		}
	}

	images, err = p.uploader.Upload(pCtx, images...)
	if err != nil {
		return fmt.Errorf("upload images: %w", err)
	}

	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}

	page := generatePage(title, urls)

	pageURL, err := p.tgAPI.CreatePage(ctx, page)
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	if err := generateOutput(pageURL, autoOpen, noDialog); err != nil {
		return fmt.Errorf("generate output: %w", err)
	}

	return nil
}

func listImages(dir string) ([]entities.MediaFile, error) {
	imageFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read img directory: %w", err)
	}

	utils.NaturalSort(imageFiles, func(e fs.DirEntry) string { return e.Name() })

	images := make([]entities.MediaFile, 0, len(imageFiles))

	for _, file := range imageFiles {
		if file.IsDir() {
			continue
		}

		if !usecases.IsImage(file.Name()) {
			continue
		}

		file := filepath.Join(dir, file.Name())

		img, err := usecases.LoadMedia(file)
		if err != nil {
			return images, fmt.Errorf("load image: %w", err)
		}

		images = append(images, img)
	}

	return images, nil
}

func generatePage(title string, imgURLs []string) entities.Page {
	res := entities.Page{
		Title:   title,
		Content: make([]entities.Node, 0, len(imgURLs)),
	}

	for _, url := range imgURLs {
		res.Content = append(res.Content, entities.Node{
			Tag:   "img",
			Attrs: map[string]string{"src": url},
		})
	}

	return res
}

func generateOutput(url string, autoOpen, silent bool) error {
	fmt.Fprintf(os.Stdout, "Article posted: %s", url)

	if silent {
		return nil
	}

	openBrowser := false
	if autoOpen {
		openBrowser = true
	} else {
		open := dialog.
			Message("Article uploaded successfully\nWould you like to open it?").
			Title("Success").YesNo()

		openBrowser = open
	}

	if openBrowser {
		if err := browser.OpenURL(url); err != nil {
			return fmt.Errorf("open url: %w", err)
		}
	}

	return nil
}
