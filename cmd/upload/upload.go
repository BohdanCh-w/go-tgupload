package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/usecases"

	whlogger "github.com/bohdanch-w/wheel/logger"
)

type uploader struct {
	logger   whlogger.Logger
	cdn      services.CDN
	parallel uint
}

func (p *uploader) upload(ctx context.Context, filePathes []string, output string, plainOutput bool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	files, err := loadFiles(filePathes)
	if err != nil {
		return fmt.Errorf("load files: %w", err)
	}

	uploader := usecases.NewCDNUploader(p.logger, p.cdn, p.parallel)

	files, err = uploader.Upload(ctx, files...)
	if err != nil {
		return fmt.Errorf("upload images: %w", err)
	}

	return generateOutput(files, output, plainOutput)
}

func loadFiles(pathes []string) ([]entities.MediaFile, error) {
	queue := make([]string, len(pathes))
	copy(queue, pathes)

	files := make([]entities.MediaFile, 0, len(pathes))

	for i := 0; i < len(queue); i++ {
		path := queue[i]

		stat, err := os.Stat(path)
		if err != nil {
			return files, fmt.Errorf("read location %q: %w", path, err)
		}

		if stat.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return files, fmt.Errorf("list directory %q: %w", path, err)
			}

			for _, entry := range entries {
				if !entry.IsDir() { // no recursion
					queue = append(queue, filepath.Join(path, entry.Name()))
				}
			}

			continue
		}

		file, err := usecases.LoadMedia(path)
		if err != nil {
			return files, fmt.Errorf("load file: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func generateOutput(files []entities.MediaFile, path string, plain bool) error {
	var w io.Writer = os.Stdout

	if len(path) != 0 {
		f, err := os.OpenFile(path, os.O_WRONLY, 0o600) // nolint: mnd
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

		data := make([]outFormat, 0, len(files))

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
