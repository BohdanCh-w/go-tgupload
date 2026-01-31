package telegraph

import (
	"context"
	"fmt"

	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
)

func Login(ctx context.Context, acc entities.Account) (string, error) {
	var (
		tgAcc = &telegraph.Account{
			AuthorURL:  acc.AuthorURL,
			AuthorName: acc.AuthorName,
			ShortName:  acc.AuthorShortName,
		}
		err error
	)

	if tgAcc.AccessToken == "" {
		tgAcc, err = telegraph.CreateAccount(*tgAcc)
		if err != nil {
			return "", fmt.Errorf("create telegraph account: %w", err)
		}
	}

	return tgAcc.AccessToken, nil
}
