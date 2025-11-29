package services

import (
	"context"

	"github.com/bohdanch-w/go-tgupload/entities"
)

type TelegraphAPI interface {
	CreatePage(ctx context.Context, page entities.Page) (string, error)
	Account(ctx context.Context, fields ...string) (entities.Account, error)
}
