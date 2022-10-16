package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	images, err := app.uploadImages(intermidiateImageData)
	if err != nil {
		log.Printf("Failed to upload images")

		return fmt.Errorf("upload images failed: %w", err)
	}

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
	if len(app.Cfg.CaptionImgPath) == 0 {
		return imglist, nil
	}

	captions := make([]string, 0, len(app.Cfg.CaptionImgPath))

	for _, path := range app.Cfg.CaptionImgPath {
		url, err := addImage(path, intermidData, images)
		if err != nil {
			return nil, err
		}

		captions = append(captions, url)
	}

	imglist = append(imglist, captions...)

	return imglist, nil
}

func (app *App) addTitleImage(intermidData map[string]string, images map[string]string, imglist []string) ([]string, error) {
	if len(app.Cfg.TitleImgPath) == 0 {
		return imglist, nil
	}

	titles := make([]string, 0, len(app.Cfg.TitleImgPath))

	for _, path := range app.Cfg.TitleImgPath {
		url, err := addImage(path, intermidData, images)
		if err != nil {
			return nil, err
		}

		titles = append(titles, url)
	}

	imglist = append(titles, imglist...)

	return imglist, nil
}

func addImage(path string, inData map[string]string, outData map[string]string) (string, error) {
	var err error

	url, ok := inData[path]
	if ok {
		return url, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read image %s: %w", path, err)
	}

	url, err = postImage(filepath.Base(path), data)
	if err != nil {
		return "", err
	}

	outData[path] = url

	return url, nil
}
