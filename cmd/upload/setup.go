package upload

import (
	"fmt"

	"github.com/urfave/cli/v2"

	wherr "github.com/bohdanch-w/wheel/errors"
	whlogger "github.com/bohdanch-w/wheel/logger"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/integrations/postimages"
	"github.com/bohdanch-w/go-tgupload/services"
)

const (
	Name         = "upload"
	logLevelFlag = "loglevel"
	outputFlag   = "output"
	apiKeyFlag   = "apiKey"
	plainFlag    = "plain"
	parallelFlag = "parallel"

	defaultParallel = 8

	ErrInvalidParams = entities.Error("invalid input params")
)

func NewCMD(logger whlogger.Logger) *cli.Command {
	return &cli.Command{
		Name:  Name,
		Usage: "upload file to CDN",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  logLevelFlag,
				Usage: "level of logging for application",
			},
			&cli.StringFlag{
				Name:  outputFlag,
				Usage: "path to saved output file",
			},
			&cli.StringFlag{
				Name:  apiKeyFlag,
				Usage: "postimage API key",
			},
			&cli.BoolFlag{
				Name:  plainFlag,
				Value: true,
				Usage: "use plain output format",
			},
			&cli.UintFlag{
				Name:    parallelFlag,
				Aliases: []string{"p"},
				Value:   defaultParallel,
				Usage:   "max parallel file uploads",
			},
		},
		Action: uploadCMD{logger: logger}.run,
	}
}

type uploadCMD struct {
	logger whlogger.Logger

	logLevel    whlogger.LogLevel
	files       []string
	output      string
	plainOutput bool
	apiKey      string
	parallel    uint
}

func (cmd uploadCMD) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithLevel(cmd.logLevel)

	if cmd.apiKey == "" {
		globalCfg, err := config.ReadConfig("")
		if err != nil {
			return fmt.Errorf("retrieve global config: %w", err)
		}

		key := globalCfg.Get(config.PostimgAPIKey)
		if key == "" {
			return wherr.Error("no api key provided")
		}

		cmd.apiKey = key
	}

	var cdn services.CDN = postimages.NewAPI(cmd.apiKey, "")

	up := uploader{
		logger:   logger,
		cdn:      cdn,
		parallel: cmd.parallel,
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

	var logLevel whlogger.LogLevel
	if err := logLevel.UnmarshalText([]byte(ctx.String(logLevelFlag))); err != nil {
		return fmt.Errorf("parse loglevel: %w", err)
	}

	cmd.logLevel = logLevel
	cmd.output = ctx.String(outputFlag)
	cmd.plainOutput = ctx.Bool(plainFlag)
	cmd.apiKey = ctx.String(apiKeyFlag)
	cmd.parallel = ctx.Uint(parallelFlag)

	return nil
}
