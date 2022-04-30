package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/browser"
	"gitlab.com/toby3d/telegraph"

	"github.com/ZUMORl/GoTeleghraphUploader/config"
	"github.com/ZUMORl/GoTeleghraphUploader/helpers"
)

type App struct {
	Cfg *config.Config
}

func (app *App) Run() error {
	var (
		log                   = app.Cfg.Logger
		intermidiateImageData map[string]string
	)

	account, err := telegraph.CreateAccount(telegraph.Account{
		AuthorName: app.Cfg.AuthorName,
		ShortName:  app.Cfg.AuthorShortName,
	})
	if err != nil {
		log.Printf("Failed to connect to telegraph: %v", err)
	}

	if app.Cfg.AuthToken != "" {
		account.AccessToken = app.Cfg.AuthToken
	}

	intermidiateImageData, err = app.loadIntermidiateImageData()
	if err != nil {
		log.Printf("intermidiate data load failed: %v", err)
	}

	images, ok := app.uploadImages(intermidiateImageData)

	if app.Cfg.IntermidDataSavePath != "" {
		if err := helpers.WriteFileJSON(app.Cfg.IntermidDataSavePath, images); err != nil {
			log.Printf("Failed to save intermidiate data")
		}
	}

	if !ok {
		log.Printf("Failed to upload images")

		return fmt.Errorf("Not all images uploaded")
	}

	html := helpers.CreateDomFromImages(images)

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

func (app *App) uploadImages(preImages map[string]string) (map[string]string, bool) {
	imageFiles, err := ioutil.ReadDir(app.Cfg.PathToImgFolder)
	if err != nil {
		app.Cfg.Logger.Fatal(err)
	}

	var (
		wg sync.WaitGroup
		ok = true
	)

	images := map[string]string{}

	for _, file := range imageFiles {
		path := filepath.Join(app.Cfg.PathToImgFolder, file.Name())

		if _, ok := preImages[path]; ok {
			images[path] = preImages[path]
			continue
		}

		wg.Add(1)

		go func(path string) {
			defer wg.Done()

			images[path], err = helpers.PostImage(path)
			if err != nil {
				ok = false
				delete(images, path)
				app.Cfg.Logger.Printf("PostImage failed with error: %v on path: %s", err, path)
			}
		}(path)
	}

	wg.Wait()
	return images, ok
}

func (app *App) loadIntermidiateImageData() (map[string]string, error) {
	if !app.Cfg.IntermidDataEnabled {
		return nil, nil
	}

	if app.Cfg.IntermidDataLoadPath == "" {
		app.Cfg.Logger.Println("Warn: Intemidiate data load enabled but path not set")

		return nil, nil
	}

	file, err := os.Open(app.Cfg.IntermidDataLoadPath)
	if err != nil {
		return nil, fmt.Errorf("Failed load intermidiate data: %v", err)
	}

	var result map[string]string
	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return nil, fmt.Errorf("Failed parsing intermidiate data: %v", err)
	}

	return result, nil
}

func (app *App) getResult(url string) (ok bool) {
	if app.Cfg.AutoOpen {
		if err := browser.OpenURL(url); err != nil {
			log.Printf("Failed to open url: %v", err)
		} else {
			ok = true
		}
	}

	if app.Cfg.PathToOutputFile != "" {
		if err := helpers.WriteFileJSON(app.Cfg.PathToOutputFile, url); err != nil {
			log.Printf("Failed to open url: %v", err)
		} else {
			ok = true
		}
	}

	if !ok {
		log.Printf("Result url: %s", url)
	}

	return ok
}
