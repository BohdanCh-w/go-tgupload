package upload

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bohdanch-w/go-tgupload/cache"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/telegraph"
)

const (
	Name         = "upload"
	logLevelFlag = "loglevel"
	cacheFlag    = "cache"
	outputFlag   = "output"
	plainFlag    = "plain"

	ErrInvalidParams = entities.Error("invalid input params")
)

func NewCMD(logger *zap.Logger) *cli.Command {
	return &cli.Command{
		Name:  Name,
		Usage: "upload file to telegraph CDN",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  logLevelFlag,
				Usage: "level of logging for application",
			},
			&cli.StringFlag{
				Name:  cacheFlag,
				Usage: "path to saved cache. If specified will use caching for CDN uploads",
			},
			&cli.StringFlag{
				Name:  outputFlag,
				Usage: "path to saved output file",
			},
			&cli.BoolFlag{
				Name:  plainFlag,
				Usage: "use plain output format",
			},
		},
		Action: uploadCMD{logger: logger}.run,
	}
}

type uploadCMD struct {
	logger *zap.Logger

	logLevel    zapcore.Level
	cache       string
	files       []string
	output      string
	plainOutput bool
}

func (cmd uploadCMD) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithOptions(zap.IncreaseLevel(cmd.logLevel))
	defer func() { _ = logger.Sync() }()

	var cdn services.CDN = telegraph.New()
	if cmd.cache != "" {
		c := cache.New(cdn, logger)

		if err := c.LoadFile(cmd.cache); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		defer func() { _ = c.SaveFile(cmd.cache) }()

		cdn = c
	}

	up := uploader{
		logger: logger.Sugar(),
		cdn:    cdn,
	}

	if err := up.upload(ctx.Context, cmd.files, cmd.output, cmd.plainOutput); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	return nil
}

func (cmd *uploadCMD) getConfig(ctx *cli.Context) error {
	cmd.files = ctx.Args().Slice()
	if len(cmd.files) == 0 {
		return entities.Error("no files specified")
	}

	logLevel, err := zapcore.ParseLevel(ctx.String(logLevelFlag))
	if err != nil {
		return fmt.Errorf("parse loglevel: %w", err)
	}

	cmd.logLevel = logLevel
	cmd.cache = ctx.String(cacheFlag)
	cmd.output = ctx.String(outputFlag)
	cmd.plainOutput = ctx.Bool(plainFlag)

	return nil
}
