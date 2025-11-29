package post

import (
	"fmt"
	"os"

	"github.com/sqweek/dialog"
	"github.com/urfave/cli/v2"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/integrations/telegraph"
	"github.com/bohdanch-w/go-tgupload/usecases"

	wherr "github.com/bohdanch-w/wheel/errors"
	whlogger "github.com/bohdanch-w/wheel/logger"
)

const (
	Name         = "post"
	logLevelFlag = "loglevel"
	cacheFlag    = "cache"
	noDialogFlag = "no-dialog"
	parallelFlag = "parallel"
	cdnFlag      = "cdn"
	titleFlag    = "title"
	browserFlag  = "browser"

	postImageAPIKeyFlag    = "post-img-key"
	awsKeyIDFlag           = "aws-key-id"
	awsSecretAccessKeyFlag = "aws-secret-access-key"
	awsRegionFlag          = "aws-region"
	awsEndpointFlag        = "aws-endpoint"
	awsS3BucketFlag        = "aws-s3-bucket"
	awsS3LocationFlag      = "aws-s3-location"
	awsS3PublicURLFlag     = "aws-s3-public-url"

	logLevelDefault = "INFO"
	parallelDefault = 8
)

func NewCMD(logger whlogger.Logger) *cli.Command { // nolint: funlen
	return &cli.Command{
		Name:  Name,
		Usage: "post telegraph article from image gallery",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  logLevelFlag,
				Usage: "level of logging for application",
				Value: logLevelDefault,
			},
			&cli.StringFlag{
				Name:  cacheFlag,
				Usage: "path to saved cache. If specified will use caching for CDN uploads",
			},
			&cli.BoolFlag{
				Name:    noDialogFlag,
				Usage:   "don't prompt window for user input",
				Aliases: []string{"s"},
			},
			&cli.UintFlag{
				Name:    parallelFlag,
				Usage:   "set number of parallel file upload",
				Aliases: []string{"p"},
				Value:   parallelDefault,
			},
			&cli.StringFlag{
				Name:  cdnFlag,
				Usage: "override preffered cdn. Supported values are ['post-image', 's3']",
			},
			&cli.BoolFlag{
				Name:    browserFlag,
				Usage:   "auto open uploaded article in the browser",
				Aliases: []string{"a"},
			},
			&cli.StringFlag{
				Name:    titleFlag,
				Usage:   "specify the title of the article. If empty, then you will be prompted later.",
				Aliases: []string{"t"},
			},
			&cli.StringFlag{
				Name:  postImageAPIKeyFlag,
				Usage: "API key for post-image CDN",
				EnvVars: []string{
					"POST_IMAGE_API_KEY",
				},
			},
			&cli.StringFlag{
				Name:   awsKeyIDFlag,
				Hidden: true,
				EnvVars: []string{
					"AWS_KEY_ID",
				},
			},
			&cli.StringFlag{
				Name:   awsSecretAccessKeyFlag,
				Hidden: true,
				EnvVars: []string{
					"AWS_SECRET_ACCESS_KEY",
				},
			},
			&cli.StringFlag{
				Name:   awsRegionFlag,
				Hidden: true,
				EnvVars: []string{
					"AWS_REGION",
				},
			},
			&cli.StringFlag{
				Name:   awsEndpointFlag,
				Hidden: true,
				EnvVars: []string{
					"AWS_ENDPOINT",
				},
			},
			&cli.StringFlag{
				Name:    awsS3BucketFlag,
				Usage:   "name of the bucket for S3 CDN",
				Aliases: []string{"bucket"},
				EnvVars: []string{
					"AWS_S3_BUCKET",
				},
			},
			&cli.StringFlag{
				Name:    awsS3LocationFlag,
				Usage:   "location in the bucket for S3 CDN",
				Aliases: []string{"location"},
				EnvVars: []string{
					"AWS_S3_LOCATION",
				},
			},
			&cli.StringFlag{
				Name:    awsS3PublicURLFlag,
				Usage:   "prefix for formed URL for S3 CDN",
				Aliases: []string{"public-url"},
				EnvVars: []string{
					"AWS_S3_PUBLIC_URL",
				},
			},
		},
		Action: postCmd{logger: logger}.run,
	}
}

type postCmd struct {
	logger    whlogger.Logger
	logLevel  whlogger.LogLevel
	cache     string
	cdn       string
	directory string
	title     string

	postImageAPIKey    string
	awsKeyID           string
	awsSecretAccessKey string
	awsRegion          string
	awsEndpoint        string
	awsS3Bucket        string
	awsS3Location      string
	awsS3PublicURL     string

	noDialog bool
	autoOpen bool
}

func (cmd postCmd) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithLevel(cmd.logLevel)

	globalCfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return fmt.Errorf("retrieve global config: %w", err)
	}

	acc := globalCfg.Account()
	if !globalCfg.Exists() || !acc.Configured() || acc.AccessToken == "" {
		return wherr.Error("account is not configured")
	}

	tg, err := telegraph.New(entities.Account{
		AuthorName:      acc.AuthorName,
		AuthorShortName: acc.AuthorShortName,
		AuthorURL:       acc.AuthorURL,
		AccessToken:     acc.AccessToken,
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	var cdnOpts usecases.CDNOptions

	cdnOpts.S3.KeyID = cmd.awsKeyID
	cdnOpts.S3.SecretAccessKey = cmd.awsSecretAccessKey
	cdnOpts.S3.Region = cmd.awsRegion
	cdnOpts.S3.Endpoint = cmd.awsEndpoint
	cdnOpts.S3.Bucket = cmd.awsS3Bucket
	cdnOpts.S3.Location = cmd.awsS3Location
	cdnOpts.S3.PublicURL = cmd.awsS3PublicURL
	cdnOpts.PostImage.APIKey = cmd.postImageAPIKey
	cdnOpts.Cache.Enable = cmd.cache != ""
	cdnOpts.Cache.FilePath = cmd.cache

	cdn, err := usecases.NewCDN(ctx.Context, cmd.cdn, globalCfg, cdnOpts)
	if err != nil {
		return fmt.Errorf("open cdn connection: %w", err)
	}

	up := poster{
		uploader: usecases.NewCDNUploader(logger, cdn, 0),
		tgAPI:    tg,
	}

	if err := up.post(ctx.Context, cmd.directory, cmd.title, cmd.noDialog, cmd.autoOpen); err != nil {
		if !cmd.noDialog {
			dialog.Message("Your article couldn't be posted due to following error:\n%s", err.Error()).Title("Error").Error()
		}

		return fmt.Errorf("post: %w", err)
	}

	return nil
}

func (cmd *postCmd) getConfig(ctx *cli.Context) error {
	cmd.directory = ctx.Args().First()
	cmd.cache = ctx.String(cacheFlag)
	cmd.title = ctx.String(titleFlag)
	cmd.noDialog = ctx.Bool(noDialogFlag)
	cmd.autoOpen = ctx.Bool(browserFlag)
	cmd.cdn = ctx.String(cdnFlag)

	cmd.postImageAPIKey = ctx.String(postImageAPIKeyFlag)
	cmd.awsKeyID = ctx.String(awsKeyIDFlag)
	cmd.awsSecretAccessKey = ctx.String(awsSecretAccessKeyFlag)
	cmd.awsRegion = ctx.String(awsRegionFlag)
	cmd.awsEndpoint = ctx.String(awsEndpointFlag)
	cmd.awsS3Bucket = ctx.String(awsS3BucketFlag)
	cmd.awsS3Location = ctx.String(awsS3LocationFlag)
	cmd.awsS3PublicURL = ctx.String(awsS3PublicURLFlag)

	var logLevel whlogger.LogLevel
	if err := logLevel.UnmarshalText([]byte(ctx.String(logLevelFlag))); err != nil {
		return fmt.Errorf("parse loglevel: %w", err)
	}

	if cmd.directory == "" {
		return wherr.Error("no source directory provided")
	} else {
		if _, err := os.Stat(cmd.directory); err != nil {
			return fmt.Errorf("verify source directory: %w", err)
		}
	}

	return nil
}
