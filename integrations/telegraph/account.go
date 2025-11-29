package telegraph

import (
	"context"
	"fmt"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/wheel/ds/hashset"
	wherr "github.com/bohdanch-w/wheel/errors"
)

func (a *API) Account(ctx context.Context, fields ...string) (entities.Account, error) {
	if len(fields) == 0 {
		return entities.Account{}, wherr.Error("no fields requested")
	}

	supportedFields := hashset.New(
		"author_name",
		"short_name",
		"author_url",
		"access_token",
		"auth_url",
		"page_count",
	)

	for _, field := range fields {
		if !supportedFields.Has(field) {
			return entities.Account{}, wherr.Errorf("%w: %q", "unsupported field", field)
		}
	}

	acc, err := a.account.GetAccountInfo(fields...)
	if err != nil {
		return entities.Account{}, fmt.Errorf("retrieve account info: %w", err)
	}

	return entities.Account{
		AuthorName:      acc.AuthorName,
		AuthorShortName: acc.ShortName,
		AuthorURL:       acc.AuthorURL,
		AccessToken:     acc.AccessToken,
		AuthURL:         acc.AuthURL,
		PageCount:       uint(acc.PageCount),
	}, nil
}
