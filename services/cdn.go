package services

import (
	"context"

	"github.com/zumorl/go-tgupload/entities"
)

type CDN interface {
	Upload(ctx context.Context, media entities.MediaFile) (string, error)
}
