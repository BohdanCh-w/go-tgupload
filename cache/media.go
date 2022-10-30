package cache

import (
	"context"
	"crypto/md5" // nolint: gosec
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
)

var _ services.CDN = (*MediaCache)(nil)

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

	var cache map[string]cachedMedia

	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("unmarshal cached data: %w", err)
	}

	for key, value := range cache {
		decoded, err := hex.DecodeString(key)
		if err != nil {
			return fmt.Errorf("decode key %s: %w", key, err)
		}

		c.cache[string(decoded)] = value
	}

	return nil
}

func (c *MediaCache) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	hash := md5.Sum(media.Data) // nolint: gosec

	if url, ok := c.getHash(hash); ok {
		return url, nil
	}

	url, err := c.retriever.Upload(ctx, media)
	if err == nil {
		c.setHash(hash, media.Path, url)
	}

	return url, err // nolint: wrapcheck
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
	cache := make(map[string]cachedMedia)
	for key, value := range c.cache {
		cache[hex.EncodeToString([]byte(key))] = value
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o666); err != nil { // nolint: gosec
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
