//go:build linux

package config

import (
	"os"
	"path/filepath"
)

func DefaultConfigLocation() (string, error) {
	return filepath.Join(os.Getenv("HOME"), ".gotg", "config.json"), nil
}
