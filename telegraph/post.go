package telegraph

import (
	"context"
	"fmt"

	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"
)

func (s *Server) CreatePage(ctx context.Context, page entities.Page) (string, error) {
	html := make([]telegraph.Node, 0, len(page.Content))
	for _, div := range page.Content {
		html = append(html, div)
	}

	p, err := s.account.CreatePage(telegraph.Page{
		Title:       page.Title,
		AuthorName:  utils.DefaultIfNil(page.AuthorName, s.account.AuthorName),
		AuthorURL:   utils.DefaultIfNil(page.AuthorURL, s.account.AuthorURL),
		Description: page.Description,
		Content:     html,
	}, true)
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}

	return p.URL, nil
}
