package post

import (
	"fmt"

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

	ErrInvalidParams = entities.Error("invalid input params")
)

func NewCMD(logger *zap.Logger) *cli.Command {
	return &cli.Command{
		Name: Name,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: logLevelFlag,
			},
			&cli.StringFlag{
				Name: cacheFlag,
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
}

func (cmd postCmd) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithOptions(zap.IncreaseLevel(cmd.logLevel))
	defer func() { _ = logger.Sync() }()

	var (
		cdn services.CDN
		tg  = telegraph.New()
	)

	cdn = tg
	if cmd.cache != "" {
		c := cache.New(tg)

		if err := c.LoadFile(cmd.cache); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		defer c.SaveFile(cmd.cache)

		cdn = c
	}

	up := poster{
		logger: logger,
		cdn:    cdn,
		tgAPI:  tg,
	}

	if err := up.post(ctx.Context, cmd.cfg); err != nil {
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

	return nil
}
