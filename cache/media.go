package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
)

var _ services.CDN = (*MediaCache)(nil)

type retrieveFunc func(context.Context, entities.MediaFile) (string, error)

type MediaCache struct {
	cache     map[string]cachedMedia
	retriever services.CDN
	mux       sync.RWMutex
}

type cachedMedia struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

func New(retriever services.CDN) *MediaCache {
	return &MediaCache{
		retriever: retriever,
		cache:     make(map[string]cachedMedia),
	}
}

func (c *MediaCache) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(data, &c.cache); err != nil {
		return fmt.Errorf("unmarshal cached data")
	}

	return nil
}

func (c *MediaCache) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	hash := md5.Sum(media.Data)

	if url, ok := c.getHash(hash); ok {
		return url, nil
	}

	url, err := c.retriever.Upload(ctx, media)
	if err == nil {
		c.setHash(hash, media.Path, url)
	}

	return url, err
}

func (c *MediaCache) getHash(hash [16]byte) (string, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	media, ok := c.cache[string(hash[:])]

	return media.URL, ok
}

func (c *MediaCache) setHash(hash [16]byte, path, url string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.cache[string(hash[:])] = cachedMedia{
		Path: path,
		URL:  url,
	}
}

func (c *MediaCache) SaveFile(path string) error {
	data, err := json.MarshalIndent(c.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o666); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
