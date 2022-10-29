package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bohdanch-w/datatypes/hashset"
	"github.com/bohdanch-w/go-tgupload/entities"
)

func LoadMedia(path string) (entities.MediaFile, error) {
	media := entities.MediaFile{
		Name: filepath.Base(path),
		Path: path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return media, fmt.Errorf("read media %s: %w", path, err)
	}

	media.Data = data

	return media, nil
}

func IsImage(path string) bool {
	return hashset.New(".png", ".jpg", ".jpeg").Has(filepath.Ext(path))
}
