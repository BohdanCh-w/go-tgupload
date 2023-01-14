package post

import (
	"fmt"

	"github.com/sqweek/dialog"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bohdanch-w/go-tgupload/cache"
	"github.com/bohdanch-w/go-tgupload/cmd/post/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/bohdanch-w/go-tgupload/telegraph"
)

const (
	Name         = "run"
	logLevelFlag = "loglevel"
	cacheFlag    = "cache"
	silentFlag   = "no-dialog"

	ErrInvalidParams = entities.Error("invalid input params")
)

func NewCMD(logger *zap.Logger) *cli.Command {
	return &cli.Command{
		Name:  Name,
		Usage: "post telegraph article according to specified config",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  logLevelFlag,
				Usage: "level of logging for application",
			},
			&cli.StringFlag{
				Name:  cacheFlag,
				Usage: "path to saved cache. If specified will use caching for CDN uploads",
			},
			&cli.BoolFlag{
				Name:    silentFlag,
				Usage:   "don't prompt window for user input",
				Aliases: []string{"s"},
			},
		},
		Action: postCmd{logger: logger}.run,
	}
}

type postCmd struct {
	logger   *zap.Logger
	cfg      config.Config
	logLevel zapcore.Level
	cache    string

	silent bool
}

func (cmd postCmd) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithOptions(zap.IncreaseLevel(cmd.logLevel))
	defer func() { _ = logger.Sync() }()

	var (
		tg               = telegraph.New()
		cdn services.CDN = tg
	)

	err := tg.Login(ctx.Context, entities.Account{
		AuthorName:      cmd.cfg.AuthorName,
		AuthorShortName: cmd.cfg.AuthorShortName,
		AuthorURL:       cmd.cfg.AuthorURL,
		AccessToken:     cmd.cfg.AuthToken,
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	if cmd.cache != "" {
		c := cache.New(tg, logger)

		if err := c.LoadFile(cmd.cache); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		defer func() { _ = c.SaveFile(cmd.cache) }()

		cdn = c
	}

	up := poster{
		logger: logger.Sugar(),
		cdn:    cdn,
		tgAPI:  tg,
	}

	if err := up.post(ctx.Context, cmd.cfg, cmd.silent); err != nil {
		if !cmd.silent {
			dialog.Message("Your article couldnt be posted due to following error:\n%s", err.Error()).Title("Error").Error()
		}

		return fmt.Errorf("post: %w", err)
	}

	return nil
}

func (cmd *postCmd) getConfig(ctx *cli.Context) error {
	path := ctx.Args().First()
	if len(path) == 0 {
		return fmt.Errorf("%w: config path is required", ErrInvalidParams)
	}

	if err := cmd.cfg.Parse(path); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	logLevel, err := zapcore.ParseLevel(ctx.String(logLevelFlag))
	if err != nil {
		return fmt.Errorf("parse loglevel: %w", err)
	}

	cmd.logLevel = logLevel
	cmd.cache = ctx.String(cacheFlag)
	cmd.silent = ctx.Bool(silentFlag)

	return nil
}
