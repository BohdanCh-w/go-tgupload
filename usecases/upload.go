package usecases

import (
	"context"
	"fmt"
	"sync"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func UploadFilesToCDN( // nolint: funlen
	ctx context.Context,
	logger *zap.SugaredLogger,
	cdn services.CDN,
	mediaFiles []entities.MediaFile,
) ([]entities.MediaFile, error) {
	var (
		sem  = semaphore.NewWeighted(8) // nolint: gomnd // TODO: make this configurable
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
	logger *zap.SugaredLogger,
	cdn services.CDN,
	mediaFile entities.MediaFile,
) (entities.MediaFile, error) {
	logger.Infof("Start %s uploading", mediaFile.Name)
	defer logger.Infof("uploading %s ended", mediaFile.Name)

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
