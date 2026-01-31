package usecases

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"

	"github.com/bohdanch-w/wheel/collections"
	wherr "github.com/bohdanch-w/wheel/errors"
	whlogger "github.com/bohdanch-w/wheel/logger"
)

const defaultCDNUploadParallel = 8

func NewCDNUploader(logger whlogger.Logger, cdn services.CDN, parallel uint) *CDNUploader {
	return &CDNUploader{
		logger:   logger,
		cdn:      cdn,
		parallel: collections.DefaultIfEmpty(parallel, defaultCDNUploadParallel),
	}
}

type CDNUploader struct {
	logger   whlogger.Logger
	cdn      services.CDN
	parallel uint
}

func (u *CDNUploader) Upload(ctx context.Context, mediaFiles ...entities.MediaFile) ([]entities.MediaFile, error) {
	if len(mediaFiles) == 0 {
		return nil, nil
	}

	if len(mediaFiles) == 1 {
		file, err := UploadFileToCDN(ctx, u.logger, u.cdn, mediaFiles[0])
		if err != nil {
			return nil, err
		}

		return []entities.MediaFile{file}, nil
	}

	return UploadFilesToCDN(ctx, u.logger, u.cdn, u.parallel, mediaFiles)
}

func UploadFilesToCDN( // nolint: funlen
	ctx context.Context,
	logger whlogger.Logger,
	cdn services.CDN,
	parallel uint,
	mediaFiles []entities.MediaFile,
) ([]entities.MediaFile, error) {
	if parallel == 0 {
		return nil, wherr.Error("invalid configuration")
	}

	var (
		sem  = semaphore.NewWeighted(int64(parallel))
		wg   sync.WaitGroup
		done = make(chan struct{})
		mErr *multierror.Error

		uploaded = make([]entities.MediaFile, len(mediaFiles))
		order    = make(map[string]int, len(mediaFiles))
		resChan  = make(chan uploadResult)
	)

	for i, val := range mediaFiles {
		order[val.Path] = i
	}

	go func() { // collector
		defer close(done)

		for res := range resChan {
			if res.err != nil {
				mErr = multierror.Append(mErr, res.err)

				continue
			}

			uploaded[order[res.media.Path]] = res.media
		}
	}()

	for i := range mediaFiles { // producer
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}

		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			defer sem.Release(1)

			res, err := UploadFileToCDN(ctx, logger, cdn, mediaFiles[idx])
			resChan <- uploadResult{
				media: res,
				err:   err,
			}
		}(i)
	}

	wg.Wait()
	close(resChan)
	<-done

	if mErr.ErrorOrNil() != nil {
		return nil, fmt.Errorf("upload images: %w", mErr)
	}

	return uploaded, nil
}

func UploadFileToCDN(
	ctx context.Context,
	logger whlogger.Logger,
	cdn services.CDN,
	mediaFile entities.MediaFile,
) (entities.MediaFile, error) {
	logger.Debugf("Start %s uploading", mediaFile.Name)
	defer logger.Debugf("uploading %s ended", mediaFile.Name)

	url, err := cdn.Upload(ctx, mediaFile)
	if err != nil {
		return mediaFile, fmt.Errorf("post image %s: %w", mediaFile.Name, err)
	}

	mediaFile.URL = url

	return mediaFile, nil
}

type uploadResult struct {
	media entities.MediaFile
	err   error
}
