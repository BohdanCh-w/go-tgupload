package telegraph

import (
	"context"
	"fmt"

	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"
)

func toNode(div entities.Node) telegraph.NodeElement {
	children := make([]telegraph.Node, 0, len(div.Children))

	for _, c := range div.Children {
		if v, ok := c.(entities.Node); ok {
			children = append(children, toNode(v))

			continue
		}

		children = append(children, c)
	}

	return telegraph.NodeElement{
		Tag:      div.Tag,
		Attrs:    div.Attrs,
		Children: children,
	}
}

func (s *Server) CreatePage(ctx context.Context, page entities.Page) (string, error) {
	html := make([]telegraph.Node, 0, len(page.Content))
	for _, div := range page.Content {
		html = append(html, toNode(div))
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
