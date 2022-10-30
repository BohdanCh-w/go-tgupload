package config

import (
	"fmt"
	"log"
	"os"

	"github.com/bohdanch-w/go-tgupload/entities"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Logger           *log.Logger
	AuthToken        string   `yaml:"auth_token"`
	PathToImgFolder  string   `yaml:"img_folder"`
	TitleImgPath     []string `yaml:"title_img_path"`
	CaptionImgPath   []string `yaml:"caption_img_path"`
	PathToOutputFile string   `yaml:"output"`
	AutoOpen         bool     `yaml:"auto_open"`

	Title           string `yaml:"title"`
	AuthorName      string `yaml:"author_name"`
	AuthorShortName string `yaml:"author_short_name"`
	AuthorURL       string `yaml:"author_url"`
}

func (c *Config) Parse(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("parse version data: %w", err)
	}

	if !c.AutoOpen && c.PathToImgFolder == "" {
		c.PathToImgFolder = "output_link.txt"
	}

	if c.AuthorShortName == "" {
		c.AuthorShortName = c.AuthorName
	}

	return c.validate()
}

func (c *Config) validate() error {
	const (
		errMissingPathToImgFolder = entities.Error("path_to_img_folder is required")
		errMissingTitle           = entities.Error("title is required")
		errMissingAuthorName      = entities.Error("author_name is required")
	)

	if c.PathToImgFolder == "" {
		return errMissingPathToImgFolder
	}

	if c.Title == "" {
		return errMissingTitle
	}

	if c.AuthorName == "" {
		return errMissingAuthorName
	}

	return nil
}
