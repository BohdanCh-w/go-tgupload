package cache

import (
	"context"
	"crypto/md5" // nolint: gosec
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/services"
	"go.uber.org/zap"
)

var _ services.CDN = (*MediaCache)(nil)

type MediaCache struct {
	retriever services.CDN
	logger    *zap.Logger

	cache map[[md5.Size]byte]cachedMedia
	mux   sync.RWMutex
}

type cachedMedia struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

func New(retriever services.CDN, logger *zap.Logger) *MediaCache {
	return &MediaCache{
		retriever: retriever,
		logger:    logger,
		cache:     make(map[[md5.Size]byte]cachedMedia),
	}
}

func (c *MediaCache) LoadFile(path string) error {
	const errCacheInvalid = entities.Error("invalid saved hash")

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(os.ErrNotExist, err) {
			c.cache = make(map[[md5.Size]byte]cachedMedia) // nullify cache

			return nil
		}

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

		if length := len(decoded); length != md5.Size {
			return fmt.Errorf("%w: hash size %d, expected %d", errCacheInvalid, length, md5.Size)
		}

		c.cache[*((*[md5.Size]byte)(decoded))] = value
	}

	return nil
}

func (c *MediaCache) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	hash := md5.Sum(media.Data) // nolint: gosec

	if cached, ok := c.getHash(hash); ok {
		if media.Path != cached.Path {
			c.logger.Warn("equal hash for different pathes", zap.String("old", cached.Path), zap.String("new", media.Path))
		}

		return cached.URL, nil
	}

	url, err := c.retriever.Upload(ctx, media)
	if err == nil {
		c.setHash(hash, media.Path, url)
	}

	return url, err // nolint: wrapcheck
}

func (c *MediaCache) getHash(hash [16]byte) (cachedMedia, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	media, ok := c.cache[hash]

	return media, ok
}

func (c *MediaCache) setHash(hash [16]byte, path, url string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cache[hash] = cachedMedia{
		Path: path,
		URL:  url,
	}
}

func (c *MediaCache) SaveFile(path string) error {
	cache := make(map[string]cachedMedia)
	for key, value := range c.cache {
		cache[hex.EncodeToString(key[:])] = value
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil { // nolint: gosec
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
