package cache

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create a mock HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Setup test helpers
func setupTestContext() ctx.CTX {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard) // Silence logs during tests
	return ctx.CTX{
		FieldLogger: logger.WithField("test", true),
	}
}

func createTestCache(t *testing.T) (ImageCache, string, *MockHTTPClient) {
	// Create temporary directory for test cache
	tmpDir, err := ioutil.TempDir("", "image-cache-test")
	assert.NoError(t, err, "Failed to create temp directory")

	// Create mock client
	mockClient := new(MockHTTPClient)

	// Create cache with test directory and mock client
	cache := &imageCache{
		cachePath:     tmpDir,
		client:        mockClient, // This works because MockHTTPClient implements the Do method as in http.Client
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
	}

	// Start worker goroutines
	for i := 0; i < 2; i++ {
		go cache.downloadWorker()
	}

	return cache, tmpDir, mockClient
}

func createMockImageResponse(imageData []byte, contentType string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(imageData)),
		Header:     http.Header{"Content-Type": {contentType}},
	}
}

func TestGetOrDownloadImage(t *testing.T) {
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()
	testURL := "http://example.com/test.jpg"
	testKey := "test-key"
	testImageData := []byte("test image data")

	// Mock HTTP client response for multiple calls (in case of background worker calls)
	mockResp := createMockImageResponse(testImageData, "image/jpeg")
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Test first download - should download the image
	filePath, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)
	assert.NoError(t, err, "GetOrDownloadImage should not error")
	assert.Contains(t, filePath, tmpDir, "File path should be within temp directory")
	assert.FileExists(t, filePath, "Image file should exist")

	// Verify file contents
	fileData, err := ioutil.ReadFile(filePath)
	assert.NoError(t, err, "Reading file should not error")
	assert.Equal(t, testImageData, fileData, "File content should match test data")

	// Test cached retrieval - should return from memory cache
	filePath2, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)
	assert.NoError(t, err, "Second call should not error")
	assert.Equal(t, filePath, filePath2, "Should return the same file path")
}

func TestCacheMapImages(t *testing.T) {
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()
	testImageData := []byte("test image data")

	// Create mock maps slice
	maps := []domain.Map{
		{
			UUID:        "map1",
			DisplayName: "Test Map 1",
			Splash:      "http://example.com/splash1.jpg",
			DisplayIcon: "http://example.com/icon1.png",
		},
		{
			UUID:        "map2",
			DisplayName: "Test Map 2",
			Splash:      "http://example.com/splash2.jpg",
		},
	}

	// Mock HTTP client responses
	mockResp1 := createMockImageResponse(testImageData, "image/jpeg")
	mockResp2 := createMockImageResponse(testImageData, "image/png")
	mockResp3 := createMockImageResponse(testImageData, "image/jpeg")

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/splash1.jpg"
	})).Return(mockResp1, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/icon1.png"
	})).Return(mockResp2, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/splash2.jpg"
	})).Return(mockResp3, nil)

	// Process the maps
	processedMaps, err := cache.CacheMapImages(testCtx, maps)
	assert.NoError(t, err, "CacheMapImages should not error")
	assert.Len(t, processedMaps, 2, "Should return same number of maps")

	// Verify the URLs have been replaced
	assert.Contains(t, processedMaps[0].Splash, "/api/cache/", "Splash URL should be replaced with cached version")
	assert.Contains(t, processedMaps[0].DisplayIcon, "/api/cache/", "Icon URL should be replaced with cached version")
	assert.Contains(t, processedMaps[1].Splash, "/api/cache/", "Splash URL should be replaced with cached version")

	// Verify that files were actually downloaded
	assert.True(t, len(cache.(*imageCache).cachedImages) > 0, "Cache should contain images")
}

func TestDownloadImageErrors(t *testing.T) {
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()
	testURL := "http://example.com/error.jpg"
	testKey := "error-key"

	// Test error case - HTTP error
	mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("HTTP error"))

	_, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)
	assert.Error(t, err, "Should return error on HTTP failure")

	// Test error case - non-OK status code
	mockClient.ExpectedCalls = nil
	mockResp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
	}
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	_, err = cache.GetOrDownloadImage(testCtx, testURL, testKey)
	assert.Error(t, err, "Should return error on non-OK status")
}

func TestPrewarmCache(t *testing.T) {
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()
	testImageData := []byte("test image data")

	urlMap := map[string]string{
		"key1": "http://example.com/img1.jpg",
		"key2": "http://example.com/img2.png",
	}

	// Mock HTTP client responses
	mockResp1 := createMockImageResponse(testImageData, "image/jpeg")
	mockResp2 := createMockImageResponse(testImageData, "image/png")

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/img1.jpg"
	})).Return(mockResp1, nil)

	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://example.com/img2.png"
	})).Return(mockResp2, nil)

	// Test prewarm
	err := cache.PrewarmCache(testCtx, urlMap)
	assert.NoError(t, err, "PrewarmCache should not error")

	// Sleep briefly to allow background downloads to complete
	time.Sleep(100 * time.Millisecond)

	// Verify files were downloaded
	// First check in-memory cache
	cacheMem := cache.(*imageCache)
	cacheMem.mutex.RLock()
	assert.Len(t, cacheMem.cachedImages, 2, "Should have 2 images in memory cache")
	cacheMem.mutex.RUnlock()

	// Then verify files exist on disk
	for _, path := range cacheMem.cachedImages {
		assert.FileExists(t, path, "Cache file should exist on disk")
	}
}
