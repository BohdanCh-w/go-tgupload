package telegraph

import (
	"context"
	"fmt"

	"gitlab.com/toby3d/telegraph"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
)

const (
	telegraphUploadAPI = "https://telegra.ph/upload"
)

var (
	_ services.CDN          = (*Server)(nil)
	_ services.TelegraphAPI = (*Server)(nil)
)

func New() *Server {
	return &Server{
		account: &telegraph.Account{},
	}
}

type Server struct {
	account *telegraph.Account
}

func (s *Server) Login(ctx context.Context, acc entities.Account) error {
	var (
		tgAcc = &telegraph.Account{
			AccessToken: acc.AccessToken,
			AuthorURL:   acc.AuthorURL,
			AuthorName:  acc.AuthorName,
			ShortName:   acc.AuthorShortName,
		}
		err error
	)

	if tgAcc.AccessToken == "" {
		tgAcc, err = telegraph.CreateAccount(*tgAcc)
		if err != nil {
			return fmt.Errorf("create telegraph account: %w", err)
		}
	}

	s.account = tgAcc

	return nil
}
