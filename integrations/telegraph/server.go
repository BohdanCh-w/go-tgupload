package telegraph

import (
	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"

	wherr "github.com/bohdanch-w/wheel/errors"
)

const (
	TelegraphAddress = "https://telegra.ph/"
)

var _ services.TelegraphAPI = (*API)(nil)

func New(acc entities.Account) (*API, error) {
	if acc.AccessToken == "" {
		return nil, wherr.Error("misconfiguration: account name and token are required")
	}

	return &API{
		account: &telegraph.Account{
			AuthorURL:   acc.AuthorURL,
			AuthorName:  acc.AuthorName,
			ShortName:   acc.AuthorShortName,
			AccessToken: acc.AccessToken,
		},
	}, nil
}

type API struct {
	account *telegraph.Account
}
