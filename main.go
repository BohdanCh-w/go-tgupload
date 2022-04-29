package main

import (
	"flag"
	"log"
	"os"

	"github.com/Bohdan-TheOne/GoTeleghraphUploader/app"
	"github.com/Bohdan-TheOne/GoTeleghraphUploader/config"
)

func main() {
	var (
		logger = log.New(os.Stdout, "", log.LstdFlags)
		config = &config.Config{Logger: logger}
	)

	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	if *configPath == "" {
		logger.Fatalln("Config file not specified")
	}

	if err := config.Parse(*configPath); err != nil {
		logger.Fatalf("Config parse failed: %v", err)
	}

	uploader := app.App{Cfg: config}
	if err := uploader.Run(); err != nil {
		logger.Fatalf("Program failed with error: %v", err)
	}

	logger.Println("Program finished work")
}
