package upload

import (
	"fmt"

	"github.com/urfave/cli/v2"

	whlogger "github.com/bohdanch-w/wheel/logger"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/usecases"
)

const (
	Name         = "upload"
	logLevelFlag = "loglevel"
	outputFlag   = "output"
	plainFlag    = "plain"
	parallelFlag = "parallel"
	cdnFlag      = "cdn"

	postImageAPIKeyFlag    = "post-img-key"
	awsKeyIDFlag           = "aws-key-id"
	awsSecretAccessKeyFlag = "aws-secret-access-key"
	awsRegionFlag          = "aws-region"
	awsEndpointFlag        = "aws-endpoint"
	awsS3BucketFlag        = "aws-s3-bucket"
	awsS3LocationFlag      = "aws-s3-location"
	awsS3PublicURLFlag     = "aws-s3-public-url"

	defaultParallel = 8
)

func NewCMD(logger whlogger.Logger) *cli.Command { // nolint: funlen
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
			&cli.StringFlag{
				Name:  cdnFlag,
				Usage: "override preffered cdn. Supported values are ['post-image', 's3']",
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
		Action: uploadCMD{logger: logger}.run,
	}
}

type uploadCMD struct {
	logger whlogger.Logger

	logLevel    whlogger.LogLevel
	files       []string
	output      string
	plainOutput bool
	parallel    uint
	cdn         string

	postImageAPIKey    string
	awsKeyID           string
	awsSecretAccessKey string
	awsRegion          string
	awsEndpoint        string
	awsS3Bucket        string
	awsS3Location      string
	awsS3PublicURL     string
}

func (cmd uploadCMD) run(ctx *cli.Context) error {
	if err := cmd.getConfig(ctx); err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	logger := cmd.logger.WithLevel(cmd.logLevel)

	globalCfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return fmt.Errorf("retrieve global config: %w", err)
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

	cdn, err := usecases.NewCDN(ctx.Context, cmd.cdn, globalCfg, cdnOpts)
	if err != nil {
		return fmt.Errorf("open cdn connection: %w", err)
	}

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
	cmd.parallel = ctx.Uint(parallelFlag)
	cmd.cdn = ctx.String(cdnFlag)

	cmd.postImageAPIKey = ctx.String(postImageAPIKeyFlag)
	cmd.awsKeyID = ctx.String(awsKeyIDFlag)
	cmd.awsSecretAccessKey = ctx.String(awsSecretAccessKeyFlag)
	cmd.awsRegion = ctx.String(awsRegionFlag)
	cmd.awsEndpoint = ctx.String(awsEndpointFlag)
	cmd.awsS3Bucket = ctx.String(awsS3BucketFlag)
	cmd.awsS3Location = ctx.String(awsS3LocationFlag)
	cmd.awsS3PublicURL = ctx.String(awsS3PublicURLFlag)

	return nil
}
