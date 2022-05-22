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

	imglist, err = app.addTitleImage(intermidiateImageData, images, imglist)
	if err != nil {
		log.Printf("post title image failed: %v", err)

		return err
	}

	imglist, err = app.addCaption(intermidiateImageData, images, imglist)
	if err != nil {
		log.Printf("post caption image failed: %v", err)

		return err
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

func (app *App) addCaption(intermidData map[string]string, images map[string]string, imglist []string) ([]string, error) {
	if app.Cfg.CaptionImgPath == "" {
		return imglist, nil
	}

	var err error

	path, ok := intermidData[app.Cfg.CaptionImgPath]
	if !ok {
		path, err = postImage(app.Cfg.CaptionImgPath)
		if err != nil {
			return nil, err
		}

		images[app.Cfg.CaptionImgPath] = path
	}

	imglist = append(imglist, path)

	return imglist, nil
}

func (app *App) addTitleImage(intermidData map[string]string, images map[string]string, imglist []string) ([]string, error) {
	if app.Cfg.TitleImgPath == "" {
		return imglist, nil
	}

	var err error

	path, ok := intermidData[app.Cfg.TitleImgPath]
	if !ok {
		path, err = postImage(app.Cfg.TitleImgPath)
		if err != nil {
			return nil, err
		}

		images[app.Cfg.TitleImgPath] = path
	}

	imglist = append([]string{path}, imglist...)

	return imglist, nil
}
