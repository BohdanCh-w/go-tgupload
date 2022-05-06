package app

import (
	"encoding/json"
	"fmt"
	"os"

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
		account, err          = app.telegraphLogin()
	)

	if err != nil {
		return err
	}

	intermidiateImageData, err = app.loadIntermidiateImageData()
	if err != nil {
		log.Printf("intermidiate data load failed: %v", err)
	}

	images, ok := app.uploadImages(intermidiateImageData)
	imglist := helpers.SortedValuesByKey(images)

	if app.Cfg.CaptionImgPath != "" {
		var path string
		var ok bool

		if path, ok = intermidiateImageData[app.Cfg.CaptionImgPath]; !ok {
			path, err = postImage(app.Cfg.CaptionImgPath)
			if err != nil {
				log.Printf("post caption image failed: %v", err)

				return err
			}

			images[app.Cfg.CaptionImgPath] = path
		}

		imglist = append(imglist, path)
	}

	if app.Cfg.IntermidDataSavePath != "" {
		save_data := helpers.JoinStringMaps(intermidiateImageData, images)

		if err := helpers.WriteFileJSON(app.Cfg.IntermidDataSavePath, save_data); err != nil {
			log.Printf("Failed to save intermidiate data")
		}
	}

	if !ok {
		log.Printf("Failed to upload images")

		return fmt.Errorf("Not all images uploaded")
	}

	html := helpers.CreateDomFromImages(imglist)

	return app.createPage(account, html)
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
