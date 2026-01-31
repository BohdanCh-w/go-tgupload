package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"

	"github.com/bohdanch-w/wheel/ds/hashset"
	wherr "github.com/bohdanch-w/wheel/errors"
)

const (
	Name = "config"

	sensitiveFlag = "sensitive"
)

func NewCMD() *cli.Command {
	return &cli.Command{
		Name:  Name,
		Usage: "manage config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: sensitiveFlag,
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:   "set",
				Action: set,
			},
		},
		Action: show,
	}
}

func set(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) < 2 { // nolint: mnd
		return wherr.Error("not enough arguments")
	}

	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	if !cfg.Exists() {
		if err := os.MkdirAll(filepath.Dir(cfg.Location), 0o600); err != nil { // nolint: mnd
			return fmt.Errorf("create config directory: %w", err)
		}
	}

	cfg.Set(args[0], strings.Join(args[1:], " "))

	if err := config.StoreConfig(cfg); err != nil {
		return fmt.Errorf("store config: %w", err)
	}

	return nil
}

func show(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	if !cfg.Exists() {
		os.Stdout.WriteString("{}")

		return nil
	}

	values := cfg.Values()

	if !ctx.Bool(sensitiveFlag) {
		sensitive := hashset.New(config.SensitiveKeys()...)

		for k, v := range values {
			if sensitive.Has(k) {
				values[k] = utils.MaskString(v)
			}
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(values); err != nil {
		return fmt.Errorf("failed to write config content: %w", err)
	}

	return nil
}
