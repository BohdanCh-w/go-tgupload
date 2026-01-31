package s3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"

	"github.com/bohdanch-w/wheel/collections"
	wherr "github.com/bohdanch-w/wheel/errors"
)

var _ services.CDN = (*MediaStorage)(nil)

func NewMediaStorage(client *s3.Client, bucket, root, publicURL string) *MediaStorage {
	return &MediaStorage{
		client:    client,
		bucket:    bucket,
		root:      collections.DefaultIfEmpty(strings.Trim(root, `/\`), "/"),
		publicURL: strings.Trim(publicURL, `/\`),
	}
}

type MediaStorage struct {
	client    *s3.Client
	bucket    string
	root      string
	publicURL string
}

func (ms *MediaStorage) Store(ctx context.Context, key string, data io.Reader) error {
	s3Obj := &s3.PutObjectInput{
		Bucket: aws.String(ms.bucket),
		Key:    aws.String(filepath.ToSlash(filepath.Join(ms.root, key))),
		Body:   data,
	}

	_, ok := data.(io.Seeker)
	if !ok {
		f, err := os.CreateTemp("", "")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}

		if _, err := io.Copy(f, data); err != nil {
			f.Close()

			return fmt.Errorf("copy data: %w", err)
		}

		f.Close()

		f, err = os.Open(f.Name())
		if err != nil {
			return fmt.Errorf("open temp file: %w", err)
		}

		defer f.Close()

		s3Obj.Body = f

		defer os.Remove(f.Name())
	}

	out, err := ms.client.PutObject(ctx, s3Obj)
	if err != nil {
		return fmt.Errorf("s3 storage: put object: %w", err)
	}

	if out == nil || out.ETag == nil {
		return wherr.Error("s3 storage: put object: no meta")
	}

	return nil
}

func (ms *MediaStorage) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	ext := filepath.Ext(media.Name)
	hash := sha256.New()

	if _, err := hash.Write(media.Data); err != nil {
		return "", fmt.Errorf("generate file hash")
	}

	key := hex.EncodeToString(hash.Sum(nil))[:32] + ext

	if err := ms.Store(ctx, key, bytes.NewReader(media.Data)); err != nil {
		return "", fmt.Errorf("store file: %w", err)
	}

	actualPath := filepath.ToSlash(filepath.Join(ms.root, key))
	url := ms.publicURL + actualPath

	return url, nil
}
