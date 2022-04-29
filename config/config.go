package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Logger           *log.Logger
	AuthToken        string `envconfig:"auth_token"          required:"true"`
	Title            string `envconfig:"title"               required:"true"`
	AuthorName       string `envconfig:"author_name"         required:"true"`
	AuthorShortName  string `envconfig:"author_short_name"`
	PathToImgFolder  string `envconfig:"img_folder"          required:"true"`
	PathToOutputFile string `envconfig:"output"`
	AutoOpen         bool   `envconfig:"auto_open"`

	IntermidDataEnabled  bool   `envconfig:"intermid_data_enabled"`
	IntermidDataSavePath string `envconfig:"intermid_data_save_path"`
	IntermidDataLoadPath string `envconfig:"intermid_data_load_path"`
}

func (c *Config) Parse(path string) error {
	if err := jsonToEnv(path); err != nil {
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

func jsonToEnv(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read failed: %v", err)
	}

	var vars map[string]interface{}
	err = json.Unmarshal(data, &vars)
	if err != nil {
		return fmt.Errorf("json parse failed: %v", err)
	}

	for key, value := range vars {
		os.Setenv(key, fmt.Sprintf("%v", value))
	}

	return nil
}
