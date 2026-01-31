//go:build windows

package config

import (
	"os"
	"path/filepath"

	wherr "github.com/bohdanch-w/wheel/errors"
)

func DefaultConfigLocation() (string, error) {
	appData, ok := os.LookupEnv("APPDATA")
	if !ok || appData == "" {
		return "", wherr.Error("APPDATA env not set")
	}

	return filepath.Join(appData, "Zumori", "go-tg", "config.json"), nil
}
