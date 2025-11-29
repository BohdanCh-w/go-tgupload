package usecases

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/integrations/postimages"
	"github.com/bohdanch-w/go-tgupload/services"
	s3storage "github.com/bohdanch-w/go-tgupload/storage/s3"

	"github.com/bohdanch-w/wheel/collections"
	wherr "github.com/bohdanch-w/wheel/errors"
)

const (
	CDNTypePostImage = "post-image"
	CDNTypeS3        = "s3"
)

type CDNOptions struct {
	S3 struct {
		KeyID           string
		SecretAccessKey string
		Region          string
		Endpoint        string
		Bucket          string
		Location        string
		PublicURL       string
	}
	PostImage struct {
		APIKey string
	}
	Cache struct {
		Enable   bool
		FilePath string
	}
}

func NewCDN(ctx context.Context, typ string, cfg config.Config, opts CDNOptions) (services.CDN, error) {
	var (
		cdn services.CDN
		err error
	)

	typ = collections.DefaultIfEmpty(typ, cfg.Get(config.PreferredCDN))

	switch typ {
	case CDNTypePostImage:
		cdn, err = newPostImageCDN(cfg, opts)
		if err != nil {
			return nil, err
		}
	case CDNTypeS3:
		cdn, err = newS3CDN(ctx, cfg, opts)
		if err != nil {
			return nil, err
		}
	case "":
		return nil, wherr.Error("cdn type is not configured")
	default:
		return nil, wherr.Errorf("%w: %q", "unsupported cdn type", typ)
	}

	if opts.Cache.Enable {
		// ...
	}

	return cdn, nil
}

func newPostImageCDN(cfg config.Config, opts CDNOptions) (*postimages.API, error) {
	postImageAPIKey := collections.DefaultIfEmpty(opts.PostImage.APIKey, cfg.Get(config.PostimgAPIKey))
	if postImageAPIKey == "" {
		return nil, wherr.Error("post-image: no api key provided")
	}

	return postimages.NewAPI(postImageAPIKey, ""), nil
}

func newS3CDN(ctx context.Context, cfg config.Config, opts CDNOptions) (*s3storage.MediaStorage, error) {
	keyID := collections.DefaultIfEmpty(opts.S3.KeyID, cfg.Get(config.AWSKeyID))
	secretKey := collections.DefaultIfEmpty(opts.S3.SecretAccessKey, cfg.Get(config.AWSSecretAccessKey))
	region := collections.DefaultIfEmpty(opts.S3.Region, cfg.Get(config.AWSRegion))
	endpoint := collections.DefaultIfEmpty(opts.S3.Endpoint, cfg.Get(config.AWSEndpoint))
	bucket := collections.DefaultIfEmpty(opts.S3.Bucket, cfg.Get(config.AWSS3Bucket))
	location := collections.DefaultIfEmpty(opts.S3.Location, cfg.Get(config.AWSS3Location))
	publicURL := collections.DefaultIfEmpty(opts.S3.PublicURL, cfg.Get(config.AWSS3PublicURL))

	if keyID == "" || secretKey == "" {
		return nil, wherr.Error("s3: missing credentials")
	}

	if bucket == "" || publicURL == "" {
		return nil, wherr.Error("s3: invalid configuration")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			keyID,
			secretKey,
			"",
		)),
		awsconfig.WithDefaultRegion(region),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("can't load AWS config: %w", err)
	}

	var s3ClientOpts []func(*s3.Options)

	if endpoint != "" {
		s3ClientOpts = append(s3ClientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	client := s3.NewFromConfig(awsCfg, s3ClientOpts...)

	return s3storage.NewMediaStorage(client, bucket, location, publicURL), nil
}
