package app

import (
	"fmt"

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

func (app *App) createPage(account *telegraph.Account, html []telegraph.Node) error {
	page, err := account.CreatePage(telegraph.Page{
		Title:      app.Cfg.Title,
		AuthorName: app.Cfg.AuthorName,
		AuthorURL:  app.Cfg.AuthorURL,
		Content:    html,
	}, true)
	if err != nil {
		return fmt.Errorf("Failed to create page: %v", err)
	}

	if ok := app.getResult(page.URL); !ok {
		return fmt.Errorf("Failed to save result")
	}

	return nil
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
			app.Cfg.Logger.Printf("Failed to save output file: %v", err)
		} else {
			ok = true
		}
	}

	app.Cfg.Logger.Printf("Article posted - url: %s", url)

	return ok
}
