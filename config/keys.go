package config

const (
	TgAuthorName       = "tg-author-name"
	TgAuthorShortName  = "tg-author-short-name"
	TgAuthorURL        = "tg-author-url"
	TgAccessToken      = "tg-access-token"
	PreferredCDN       = "preferred-cdn"
	PostimgAPIKey      = "postimg-api-key"
	AWSKeyID           = "aws-key-id"
	AWSSecretAccessKey = "aws-secret-access-key"
	AWSRegion          = "aws-region"
	AWSEndpoint        = "aws-endpoint"
	AWSS3Bucket        = "aws-s3-bucket"
	AWSS3Location      = "aws-s3-location"
	AWSS3PublicURL     = "aws-s3-public-url"
)

func SensitiveKeys() []string {
	return []string{
		TgAccessToken,
		PostimgAPIKey,
		AWSKeyID,
		AWSSecretAccessKey,
	}
}
