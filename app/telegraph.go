package app

import (
	"github.com/ZUMORl/GoTeleghraphUploader/helpers"
	"github.com/pkg/browser"
	"gitlab.com/toby3d/telegraph"
)

func (app *App) telegraphLogin() (*telegraph.Account, error) {
	account, err := telegraph.CreateAccount(telegraph.Account{
		AuthorName: app.Cfg.AuthorName,
		ShortName:  app.Cfg.AuthorShortName,
	})

	if err != nil {
		app.Cfg.Logger.Printf("Failed to connect to telegraph: %v", err)

		return nil, err
	}

	if app.Cfg.AuthToken != "" {
		account.AccessToken = app.Cfg.AuthToken
	}

	return account, err
}

func (app *App) getResult(url string) (ok bool) {
	if app.Cfg.AutoOpen {
		if err := browser.OpenURL(url); err != nil {
			app.Cfg.Logger.Printf("Failed to open url: %v", err)
		} else {
			ok = true
		}
	}

	if app.Cfg.PathToOutputFile != "" {
		if err := helpers.WriteFileJSON(app.Cfg.PathToOutputFile, url); err != nil {
			app.Cfg.Logger.Printf("Failed to open url: %v", err)
		} else {
			ok = true
		}
	}

	if !ok {
		app.Cfg.Logger.Printf("Result url: %s", url)
	}

	return ok
}
