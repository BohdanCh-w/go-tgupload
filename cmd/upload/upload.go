package post

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/usecases"
)

type uploader struct {
	logger *zap.Logger
	cdn    services.CDN
}

func (p *uploader) upload(ctx context.Context, cfg config) error {
	pCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	files, err := loadFiles(cfg.files)
	if err != nil {
		return fmt.Errorf("load files: %w", err)
	}

	files, err = usecases.UploadFilesToCDN(pCtx, p.cdn, files)
	if err != nil {
		return fmt.Errorf("upload images: %w", err)
	}

	return generateOutput(files, cfg.output, cfg.plainOutput)
}

func loadFiles(pathes []string) ([]entities.MediaFile, error) {
	files := make([]entities.MediaFile, len(pathes))

	for _, file := range pathes {
		file, err := usecases.LoadMedia(file)
		if err != nil {
			return files, fmt.Errorf("load image: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func generateOutput(files []entities.MediaFile, path string, plain bool) error {
	var w io.Writer = os.Stdout

	if len(path) != 0 {
		f, err := os.OpenFile(path, os.O_WRONLY, 0o666)
		if err != nil {
			return fmt.Errorf("open output file: %w", err)
		}
		defer f.Close()

		w = f
	}

	if plain {
		for _, file := range files {
			if _, err := w.Write([]byte(file.URL + "\n")); err != nil {
				return fmt.Errorf("write data: %w", err)
			}
		}
	} else {
		type outFormat struct {
			Path string `json:"path"`
			URL  string `json:"url"`
		}

		data := make([]outFormat, len(files))

		for _, file := range files {
			data = append(data, outFormat{
				Path: file.Path,
				URL:  file.URL,
			})
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)

		if err := enc.Encode(data); err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
	}

	return nil
}
