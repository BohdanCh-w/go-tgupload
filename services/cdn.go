package services

import (
	"context"

	"github.com/bohdanch-w/go-tgupload/entities"
)

type CDN interface {
	Upload(ctx context.Context, media entities.MediaFile) (string, error)
}
