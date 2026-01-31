package telegraph

import (
	"context"
	"fmt"

	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
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

func (a *API) CreatePage(ctx context.Context, page entities.Page) (string, error) {
	html := make([]telegraph.Node, 0, len(page.Content))
	for _, div := range page.Content {
		html = append(html, toNode(div))
	}

	p, err := a.account.CreatePage(telegraph.Page{
		Title:       page.Title,
		AuthorName:  a.account.AuthorName,
		AuthorURL:   a.account.AuthorURL,
		Description: page.Description,
		Content:     html,
	}, true)
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}

	return p.URL, nil
}
