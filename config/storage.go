package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/bohdanch-w/wheel/collections"
	orderedset "github.com/bohdanch-w/wheel/ds/ordered-set"
	wherr "github.com/bohdanch-w/wheel/errors"
)

const defaultProfileName = "default"

type configReadOptions struct {
	locationRetriever func() (string, error)
}

type ConfigOption func(*configReadOptions)

func WithConfigLocation(s string) ConfigOption {
	return func(cro *configReadOptions) {
		cro.locationRetriever = func() (string, error) {
			return s, nil
		}
	}
}

func DefaultConfigLocation() (string, error) {
	appData, ok := os.LookupEnv("APPDATA")
	if !ok || appData == "" {
		return "", wherr.Error("APPDATA env not set")
	}

	return filepath.Join(appData, "Zumori", "go-tg", "config.json"), nil
}

func ReadConfig(profile string, opts ...ConfigOption) (Config, error) {
	options := configReadOptions{
		locationRetriever: DefaultConfigLocation,
	}

	for _, opt := range opts {
		opt(&options)
	}

	path, err := options.locationRetriever()
	if err != nil {
		return Config{}, fmt.Errorf("get config location: %w", err)
	}

	profile = collections.DefaultIfEmpty(profile, defaultProfileName)

	return readConfig(path, profile)
}

func readConfig(path, profile string) (Config, error) {
	config, err := readRawConfig(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{
				Location: path,
				Profile:  defaultProfileName,
			}, nil
		}

		return Config{}, err
	}

	values := config.Profiles[profile]
	if values == nil {
		values = make(map[string]string)
	}

	return Config{
		Profile:  profile,
		Location: path,
		values:   values,
	}, nil
}

func ListProfiles(opts ...ConfigOption) ([]string, error) {
	options := configReadOptions{
		locationRetriever: DefaultConfigLocation,
	}

	for _, opt := range opts {
		opt(&options)
	}

	path, err := options.locationRetriever()
	if err != nil {
		return nil, fmt.Errorf("get config location: %w", err)
	}

	config, err := readRawConfig(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{defaultProfileName}, nil
		}

		return nil, err
	}

	profiles := orderedset.New(defaultProfileName)
	profiles.Add(slices.Collect(maps.Keys(config.Profiles))...)

	return profiles.Values(), nil
}

func readRawConfig(path string) (storedConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return storedConfig{}, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	var config storedConfig

	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return storedConfig{}, fmt.Errorf("config file content is invalid: %w", err)
	}

	return config, nil
}

func StoreConfig(cfg Config) error {
	if cfg.Location == "" || cfg.Profile == "" {
		return wherr.Error("config not initialized")
	}

	config, err := readRawConfig(cfg.Location)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(cfg.Location), 0o600); err != nil { // nolint: mnd
			return fmt.Errorf("create config directory: %w", err)
		}
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]map[string]string)
	}

	config.Profiles[cfg.Profile] = cfg.values

	return storeRawConfig(config, cfg.Location)
}

func storeRawConfig(config storedConfig, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o600) // nolint: mnd
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(config); err != nil {
		return fmt.Errorf("failed to write config content: %w", err)
	}

	return nil
}

type storedConfig struct {
	Profiles map[string]map[string]string `json:"profiles"`
}
