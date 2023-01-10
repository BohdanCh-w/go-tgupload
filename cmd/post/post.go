package post

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	"go.uber.org/zap"

	"github.com/bohdanch-w/go-tgupload/cmd/post/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/usecases"
)

type poster struct {
	logger *zap.SugaredLogger
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

	images, err = usecases.UploadFilesToCDN(pCtx, p.logger, p.cdn, images)
	if err != nil {
		return fmt.Errorf("upload images: %w", err)
	}

	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}

	page := generatePage(cfg.Title, cfg.AuthorName, cfg.AuthorURL, urls)

	pageURL, err := p.tgAPI.CreatePage(ctx, page)
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	return generateOutput(pageURL, cfg.PathToOutputFile, cfg.AutoOpen)
}

func listImages(dir string, titles, captions []string) ([]entities.MediaFile, error) {
	imageFiles, err := os.ReadDir(dir)
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
		if err := os.WriteFile(outputPath, []byte(url), 0o666); err != nil { // nolint: gosec
			return fmt.Errorf("write file: %w", err)
		}
	}

	return nil
}
