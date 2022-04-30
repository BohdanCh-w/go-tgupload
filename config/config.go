package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Logger           *log.Logger
	AuthToken        string `envconfig:"auth_token"`
	PathToImgFolder  string `envconfig:"img_folder"          required:"true"`
	PathToOutputFile string `envconfig:"output"`
	AutoOpen         bool   `envconfig:"auto_open"`

	Title           string `envconfig:"title"               required:"true"`
	AuthorName      string `envconfig:"author_name"         required:"true"`
	AuthorShortName string `envconfig:"author_short_name"`
	AuthorURL       string `envconfig:"author_url"`

	IntermidDataEnabled  bool   `envconfig:"intermid_data_enabled"`
	IntermidDataSavePath string `envconfig:"intermid_data_save_path"`
	IntermidDataLoadPath string `envconfig:"intermid_data_load_path"`
}

func (c *Config) Parse(path string) error {
	if err := yamlToEnv(path); err != nil {
		return fmt.Errorf("config parse failed: %v", err)
	}

	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("process env: %w", err)
	}

	if !c.AutoOpen && c.PathToImgFolder == "" {
		c.PathToImgFolder = "output_link.txt"
	}

	if c.AuthorShortName == "" {
		c.AuthorShortName = c.AuthorName
	}

	if c.IntermidDataEnabled && c.IntermidDataSavePath == "" {
		c.IntermidDataSavePath = c.PathToOutputFile
	}

	return nil
}

func yamlToEnv(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read failed: %v", err)
	}

	var vars map[string]interface{}
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return fmt.Errorf("parse version data: %w", err)
	}

	for key, value := range vars {
		os.Setenv(key, fmt.Sprintf("%v", value))
	}

	return nil
}
