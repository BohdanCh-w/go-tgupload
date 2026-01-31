package config

import (
	"maps"

	"github.com/bohdanch-w/go-tgupload/entities"
)

type Config struct {
	Location string
	Profile  string
	values   map[string]string
}

func (cfg *Config) Set(key, value string) {
	if cfg.values == nil {
		cfg.values = make(map[string]string)
	}

	cfg.values[key] = value
}

func (cfg *Config) GetOK(key string) (string, bool) {
	if cfg == nil || cfg.values == nil {
		return "", false
	}

	k, ok := cfg.values[key]

	return k, ok
}

func (cfg *Config) Get(key string) string {
	v, _ := cfg.GetOK(key)

	return v
}

func (cfg *Config) Exists() bool {
	return cfg.values != nil
}

func (cfg *Config) Values() map[string]string {
	v := make(map[string]string)
	maps.Copy(v, cfg.values)

	return v
}

func (cfg *Config) Account() entities.Account {
	return entities.Account{
		AuthorName:      cfg.values[TgAuthorName],
		AuthorShortName: cfg.values[TgAuthorShortName],
		AuthorURL:       cfg.values[TgAuthorURL],
		AccessToken:     cfg.values[TgAccessToken],
	}
}

func (cfg *Config) SetAccount(acc entities.Account) {
	cfg.values[TgAuthorName] = acc.AuthorName
	cfg.values[TgAuthorShortName] = acc.AuthorShortName
	cfg.values[TgAuthorURL] = acc.AuthorURL
	cfg.values[TgAccessToken] = acc.AccessToken
}
