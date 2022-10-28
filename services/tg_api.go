package services

import (
	"context"

	"github.com/zumorl/go-tgupload/entities"
)

type TelegraphAPI interface {
	CreatePage(ctx context.Context, page entities.Page) (string, error)
}
