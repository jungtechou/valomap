package cache

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jungtechou/valomap/config"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
)

// Variable to allow mocking of os.MkdirAll
var osMkdirAll = os.MkdirAll

// HTTPClient is an interface for the http.Client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ImageCache is responsible for caching map images
type ImageCache interface {
	// GetOrDownloadImage retrieves an image from the cache or downloads it if not found
	GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error)

	// PrewarmCache downloads and caches all images in the provided URL map
	PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error

	// CacheMapImages processes maps to replace image URLs with cached versions
	CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error)

	// Shutdown gracefully shuts down the cache service
	Shutdown()
}

type imageCache struct {
	cachePath     string
	client        HTTPClient
	mutex         sync.RWMutex
	cachedImages  map[string]string
	downloadQueue chan downloadTask
	wg            sync.WaitGroup
}

type downloadTask struct {
	URL      string
	CacheKey string
	Ctx      ctx.CTX
}

// NewImageCache creates a new image cache service
func NewImageCache(cfg *config.Config, client *http.Client) (ImageCache, error) {
	// Default cache path is ./images-cache relative to the server
	cachePath := "/home/appuser/images-cache"

	// If testing, use a temporary directory
	if cfg != nil && cfg.Server.Port == "test" {
		var err error
		cachePath, err = ioutil.TempDir("", "image-cache-test")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	// Create cache directory if it doesn't exist
	if err := osMkdirAll(cachePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &imageCache{
		cachePath:     cachePath,
		client:        client,
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
	}

	// Start worker goroutines to handle downloads in the background
	for i := 0; i < 3; i++ {
		go cache.downloadWorker()
	}

	return cache, nil
}

// downloadWorker processes image download requests from the queue
func (c *imageCache) downloadWorker() {
	for task := range c.downloadQueue {
		c.downloadImage(task.Ctx, task.URL, task.CacheKey)
		c.wg.Done()
	}
}

// GetOrDownloadImage gets an image from the cache or downloads it if not found
func (c *imageCache) GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	// Check if the image is already in memory cache
	c.mutex.RLock()
	cachedPath, exists := c.cachedImages[cacheKey]
	c.mutex.RUnlock()

	if exists {
		ctx.FieldLogger.WithFields(logrus.Fields{
			"cache_key": cacheKey,
			"path":      cachedPath,
		}).Debug("Image found in memory cache")
		return cachedPath, nil
	}

	// Check if the image is in the file system cache
	filePath := filepath.Join(c.cachePath, cacheKey+filepath.Ext(imageURL))
	if _, err := os.Stat(filePath); err == nil {
		// Image exists in filesystem, add to memory cache
		c.mutex.Lock()
		c.cachedImages[cacheKey] = filePath
		c.mutex.Unlock()

		ctx.FieldLogger.WithFields(logrus.Fields{
			"cache_key": cacheKey,
			"path":      filePath,
		}).Debug("Image found in filesystem cache")
		return filePath, nil
	}

	// Image not found in cache, download it synchronously
	return c.downloadImage(ctx, imageURL, cacheKey)
}

// downloadImage downloads an image and stores it in the cache
func (c *imageCache) downloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	ctx.FieldLogger.WithFields(logrus.Fields{
		"url":       imageURL,
		"cache_key": cacheKey,
	}).Info("Downloading image")

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add optimization headers for faster downloads
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Determine file extension from URL or content type
	extension := filepath.Ext(imageURL)
	if extension == "" {
		// Try to determine from content type
		contentType := resp.Header.Get("Content-Type")
		switch contentType {
		case "image/jpeg":
			extension = ".jpg"
		case "image/png":
			extension = ".png"
		case "image/gif":
			extension = ".gif"
		case "image/webp":
			extension = ".webp"
		default:
			extension = ".jpg" // Default to jpg if unknown
		}
	}

	// Create output file path
	filePath := filepath.Join(c.cachePath, cacheKey+extension)

	// Create output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy response body to file
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		os.Remove(filePath) // Clean up partial file
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Store in memory cache
	c.mutex.Lock()
	c.cachedImages[cacheKey] = filePath
	c.mutex.Unlock()

	ctx.FieldLogger.WithFields(logrus.Fields{
		"url":       imageURL,
		"cache_key": cacheKey,
		"path":      filePath,
	}).Info("Image downloaded and cached successfully")

	return filePath, nil
}

// PrewarmCache downloads and caches all images in the provided URL map
func (c *imageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	ctx.FieldLogger.WithField("image_count", len(urlMap)).Info("Prewarming image cache")

	startTime := time.Now()

	// Queue all downloads
	for cacheKey, url := range urlMap {
		// Check if already cached to avoid unnecessary downloads
		c.mutex.RLock()
		_, exists := c.cachedImages[cacheKey]
		c.mutex.RUnlock()

		if !exists {
			// Check if image exists in filesystem
			filePath := filepath.Join(c.cachePath, cacheKey+filepath.Ext(url))
			if _, err := os.Stat(filePath); err == nil {
				// Image exists in filesystem, add to memory cache
				c.mutex.Lock()
				c.cachedImages[cacheKey] = filePath
				c.mutex.Unlock()
				continue
			}

			// Queue the download task
			c.wg.Add(1)
			c.downloadQueue <- downloadTask{
				URL:      url,
				CacheKey: cacheKey,
				Ctx:      ctx,
			}
		}
	}

	// Wait for all downloads to complete with a timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		duration := time.Since(startTime)
		ctx.FieldLogger.WithFields(logrus.Fields{
			"duration_ms": duration.Milliseconds(),
			"image_count": len(urlMap),
		}).Info("Cache prewarm completed")
		return nil
	case <-time.After(60 * time.Second):
		ctx.FieldLogger.Warn("Cache prewarm timed out after 60 seconds")
		return fmt.Errorf("cache prewarm timed out after 60 seconds")
	}
}

// CacheMapImages processes maps and replaces image URLs with cached versions
func (c *imageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	if len(maps) == 0 {
		ctx.FieldLogger.Warn("No maps to process for caching")
		return maps, nil
	}

	ctx.FieldLogger.WithField("map_count", len(maps)).Info("Processing maps for image caching")

	for i := range maps {
		mapPtr := &maps[i]
		if mapPtr.UUID == "" {
			ctx.FieldLogger.Warn("Found invalid map entry, skipping")
			continue
		}

		cacheKey := "map_" + mapPtr.UUID

		// Cache splash image if available
		if mapPtr.Splash != "" {
			splashKey := cacheKey + "_splash"
			cachedPath, err := c.GetOrDownloadImage(ctx, mapPtr.Splash, splashKey)
			if err == nil {
				mapPtr.Splash = "/api/cache/" + filepath.Base(cachedPath)
			} else {
				ctx.FieldLogger.WithFields(logrus.Fields{
					"url":   mapPtr.Splash,
					"error": err,
				}).Warn("Failed to cache splash image")
			}
		}

		// Cache display icon if available
		if mapPtr.DisplayIcon != "" {
			iconKey := cacheKey + "_icon"
			cachedPath, err := c.GetOrDownloadImage(ctx, mapPtr.DisplayIcon, iconKey)
			if err == nil {
				mapPtr.DisplayIcon = "/api/cache/" + filepath.Base(cachedPath)
			} else {
				ctx.FieldLogger.WithFields(logrus.Fields{
					"url":   mapPtr.DisplayIcon,
					"error": err,
				}).Warn("Failed to cache display icon")
			}
		}

		// Additional images could be cached here if needed
	}

	return maps, nil
}

// Shutdown gracefully shuts down the cache service
func (c *imageCache) Shutdown() {
	// Close the download queue to prevent new tasks
	close(c.downloadQueue)

	// Wait for all pending downloads to complete
	c.wg.Wait()
}
