package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jungtechou/valomap/config"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Create a patch for os.MkdirAll
var originalOsMkdirAll = osMkdirAll

func patchOsMkdirAll(path string, perm os.FileMode) error {
	// For tests, return success for any path
	return nil
}

// MockHTTPClient2 is another HTTP client mock for testing
type MockHTTPClient2 struct {
	mock.Mock
}

func (m *MockHTTPClient2) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// TestCacheMapImages_Coverage tests the CacheMapImages function for better coverage
func TestCacheMapImages_Coverage(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard) // Silence logs
	testCtx := ctx.CTX{
		FieldLogger: logrus.NewEntry(logger),
		Context:     context.Background(),
	}

	// Create temp directory for cache
	tempDir, err := ioutil.TempDir("", "cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock HTTP client
	mockClient := new(MockHTTPClient2)

	// Setup successful response for image download
	successResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("fake image data"))),
		Header:     make(http.Header),
	}
	successResp.Header.Set("Content-Type", "image/jpeg")

	// Setup error response
	errorResp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader([]byte("not found"))),
		Header:     make(http.Header),
	}

	// Configure mock client behavior
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/splash1.jpg"
	})).Return(successResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/icon1.jpg"
	})).Return(successResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/splash2.jpg"
	})).Return(errorResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/icon2.jpg"
	})).Return(nil, fmt.Errorf("network error"))

	// Create image cache
	cache := &imageCache{
		cachePath:     tempDir,
		client:        mockClient,
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
		quit:          make(chan struct{}),
	}

	// Start worker
	go cache.downloadWorker()

	// Test with valid maps
	testMaps := []domain.Map{
		{
			UUID:        "map1",
			DisplayName: "Map One",
			Splash:      "https://example.com/splash1.jpg",
			DisplayIcon: "https://example.com/icon1.jpg",
		},
		{
			UUID:        "map2",
			DisplayName: "Map Two",
			Splash:      "https://example.com/splash2.jpg", // Will 404
			DisplayIcon: "https://example.com/icon2.jpg",   // Will error
		},
		{
			UUID:        "", // Invalid map entry
			DisplayName: "Invalid Map",
		},
	}

	// Process maps
	resultMaps, err := cache.CacheMapImages(testCtx, testMaps)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(resultMaps))

	// Successful downloads should have updated URLs
	assert.Contains(t, resultMaps[0].Splash, "/api/cache/")
	assert.Contains(t, resultMaps[0].DisplayIcon, "/api/cache/")

	// Failed downloads should retain original URLs
	assert.Equal(t, "https://example.com/splash2.jpg", resultMaps[1].Splash)
	assert.Equal(t, "https://example.com/icon2.jpg", resultMaps[1].DisplayIcon)

	// Shutdown the cache
	cache.Shutdown()

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestDownloadImage_Coverage tests the downloadImage function with different response types
func TestDownloadImage_Coverage(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard) // Silence logs
	testCtx := ctx.CTX{
		FieldLogger: logrus.NewEntry(logger),
		Context:     context.Background(),
	}

	// Create temp directory for cache
	tempDir, err := ioutil.TempDir("", "cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock HTTP client
	mockClient := new(MockHTTPClient2)

	// Setup responses with different content types
	jpegResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("jpeg image data"))),
		Header:     make(http.Header),
	}
	jpegResp.Header.Set("Content-Type", "image/jpeg")

	pngResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("png image data"))),
		Header:     make(http.Header),
	}
	pngResp.Header.Set("Content-Type", "image/png")

	webpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("webp image data"))),
		Header:     make(http.Header),
	}
	webpResp.Header.Set("Content-Type", "image/webp")

	gifResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("gif image data"))),
		Header:     make(http.Header),
	}
	gifResp.Header.Set("Content-Type", "image/gif")

	unknownResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("unknown image data"))),
		Header:     make(http.Header),
	}
	unknownResp.Header.Set("Content-Type", "application/octet-stream")

	errorResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader([]byte("server error"))),
		Header:     make(http.Header),
	}

	// Configure mock client behavior
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image.jpg"
	})).Return(jpegResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image.png"
	})).Return(pngResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image.webp"
	})).Return(webpResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image.gif"
	})).Return(gifResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image.bin"
	})).Return(unknownResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/error"
	})).Return(errorResp, nil)

	// Create image cache
	cache := &imageCache{
		cachePath:    tempDir,
		client:       mockClient,
		cachedImages: make(map[string]string),
	}

	// Test with different image types
	imagePath, err := cache.downloadImage(testCtx, "https://example.com/image.jpg", "jpg_test")
	assert.NoError(t, err)
	assert.Contains(t, imagePath, ".jpg")

	imagePath, err = cache.downloadImage(testCtx, "https://example.com/image.png", "png_test")
	assert.NoError(t, err)
	assert.Contains(t, imagePath, ".png")

	imagePath, err = cache.downloadImage(testCtx, "https://example.com/image.webp", "webp_test")
	assert.NoError(t, err)
	assert.Contains(t, imagePath, ".webp")

	imagePath, err = cache.downloadImage(testCtx, "https://example.com/image.gif", "gif_test")
	assert.NoError(t, err)
	assert.Contains(t, imagePath, ".gif")

	imagePath, err = cache.downloadImage(testCtx, "https://example.com/image.bin", "unknown_test")
	assert.NoError(t, err)
	assert.Contains(t, imagePath, ".bin") // Extension should be based on the URL

	// Test error case
	imagePath, err = cache.downloadImage(testCtx, "https://example.com/error", "error_test")
	assert.Error(t, err)
	assert.Empty(t, imagePath)

	// Verify expectations
	mockClient.AssertExpectations(t)
}

// TestNewImageCache_Coverage tests the NewImageCache function
func TestNewImageCache_Coverage(t *testing.T) {
	// Save original and patch it
	oldMkdirAll := osMkdirAll
	osMkdirAll = patchOsMkdirAll
	defer func() {
		osMkdirAll = oldMkdirAll
	}()

	// Test with nil config
	cache, err := NewImageCache(nil, &http.Client{})
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	// Test with test port
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "test",
		},
	}
	cache, err = NewImageCache(cfg, &http.Client{})
	assert.NoError(t, err)
	assert.NotNil(t, cache)
}

// TestPrewarmCache_Coverage tests the PrewarmCache function to improve coverage
func TestPrewarmCache_Coverage(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard) // Silence logs
	testCtx := ctx.CTX{
		FieldLogger: logrus.NewEntry(logger),
		Context:     context.Background(),
	}

	// Create temp directory for cache
	tempDir, err := ioutil.TempDir("", "prewarm-coverage-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock HTTP client
	mockClient := new(MockHTTPClient2)

	// Setup successful response for image download
	successResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("fake image data"))),
		Header:     make(http.Header),
	}
	successResp.Header.Set("Content-Type", "image/jpeg")

	// Configure mock client behavior for multiple URLs
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image1.jpg"
	})).Return(successResp, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/image2.png"
	})).Return(successResp, nil)

	// Setup error for one URL
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://example.com/error.jpg"
	})).Return(nil, fmt.Errorf("network error"))

	// Create cache
	cache := &imageCache{
		cachePath:     tempDir,
		client:        mockClient,
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
		quit:          make(chan struct{}),
	}

	// Start worker goroutines
	for i := 0; i < 2; i++ {
		go cache.downloadWorker()
	}

	// Create URL map with file that already exists to test filesystem cache detection
	existingFile := filepath.Join(tempDir, "existing.jpg")
	err = ioutil.WriteFile(existingFile, []byte("existing image"), 0644)
	require.NoError(t, err)

	urlMap := map[string]string{
		"image1":    "https://example.com/image1.jpg",
		"image2":    "https://example.com/image2.png",
		"error":     "https://example.com/error.jpg",
		"existing":  "https://example.com/existing.jpg",
	}

	// Test PrewarmCache
	err = cache.PrewarmCache(testCtx, urlMap)
	require.NoError(t, err)

	// Allow background downloads to complete
	time.Sleep(100 * time.Millisecond)
	cache.wg.Wait()

	// Verify images were cached
	cache.mutex.RLock()
	assert.Contains(t, cache.cachedImages, "image1")
	assert.Contains(t, cache.cachedImages, "image2")
	assert.Contains(t, cache.cachedImages, "existing")
	cache.mutex.RUnlock()

	// Shutdown cache
	cache.Shutdown()

	// Verify expectations
	mockClient.AssertExpectations(t)
}
