package cache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jungtechou/valomap/config"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Constants for tests
const testPrewarmTimeout = 1 * time.Second

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
	logger.SetOutput(io.Discard) // Silence logs during tests
	return ctx.CTX{
		FieldLogger: logger.WithField("test", true),
	}
}

func createTestCache(t *testing.T) (ImageCache, string, *MockHTTPClient) {
	// Create temporary directory for test cache
	tmpDir, err := os.MkdirTemp("", "image-cache-test")
	assert.NoError(t, err, "Failed to create temp directory")

	// Create mock client
	mockClient := new(MockHTTPClient)

	// Create cache with test directory and mock client
	cache := &imageCache{
		cachePath:     tmpDir,
		client:        mockClient, // This works because MockHTTPClient implements the Do method as in http.Client
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
		quit:          make(chan struct{}), // Add quit channel for clean shutdown
	}

	// Start worker goroutines
	for i := 0; i < 2; i++ {
		go cache.downloadWorker()
	}

	// Add test cleanup to ensure goroutines are terminated, but only do it once
	var once sync.Once
	t.Cleanup(func() {
		once.Do(func() {
			cache.Shutdown()
		})
		// Always clean up the temp directory
		os.RemoveAll(tmpDir)
	})

	return cache, tmpDir, mockClient
}

func createMockImageResponse(imageData []byte, contentType string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(imageData)),
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
	fileData, err := os.ReadFile(filePath)
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
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	_, err = cache.GetOrDownloadImage(testCtx, testURL, testKey)
	assert.Error(t, err, "Should return error on non-OK status")
}

func TestPrewarmCache(t *testing.T) {
	// Test 1: Successful prewarming
	t.Run("Successful prewarming", func(t *testing.T) {
		cache, tmpDir, mockClient := createTestCache(t)
		defer os.RemoveAll(tmpDir)
		// Shutdown will be called by t.Cleanup

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

		// Run prewarming
		err := cache.PrewarmCache(testCtx, urlMap)

		// Assertions
		assert.NoError(t, err, "PrewarmCache should not error")

		// Verify both images were downloaded
		mockClient.AssertNumberOfCalls(t, "Do", 2)

		// Verify both images are in memory cache
		cacheMem := cache.(*imageCache)
		cacheMem.mutex.RLock()
		assert.Len(t, cacheMem.cachedImages, 2, "Should have 2 images in memory cache")
		assert.Contains(t, cacheMem.cachedImages, "key1", "First key should be in memory cache")
		assert.Contains(t, cacheMem.cachedImages, "key2", "Second key should be in memory cache")
		cacheMem.mutex.RUnlock()
	})

	// Test 2: Prewarming with existing cache
	t.Run("Prewarming with existing cache", func(t *testing.T) {
		cache, tmpDir, mockClient := createTestCache(t)
		defer os.RemoveAll(tmpDir)
		// Shutdown will be called by t.Cleanup

		testCtx := setupTestContext()
		testImageData := []byte("test image data")

		// Pre-cache one image
		key1 := "key1"
		url1 := "http://example.com/img1.jpg"

		// Create a file directly to simulate pre-cached image
		filePath := filepath.Join(tmpDir, key1+".jpg")
		err := os.WriteFile(filePath, testImageData, 0644)
		assert.NoError(t, err, "Failed to write test file")

		// Add to memory cache
		cacheMem := cache.(*imageCache)
		cacheMem.mutex.Lock()
		cacheMem.cachedImages[key1] = filePath
		cacheMem.mutex.Unlock()

		// URL map with both pre-cached and new image
		urlMap := map[string]string{
			key1:   url1,
			"key2": "http://example.com/img2.png",
		}

		// Mock HTTP client response for the second image only
		mockResp := createMockImageResponse(testImageData, "image/png")
		mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://example.com/img2.png"
		})).Return(mockResp, nil)

		// Run prewarming
		err = cache.PrewarmCache(testCtx, urlMap)

		// Assertions
		assert.NoError(t, err, "PrewarmCache should not error")

		// Verify only the non-cached image was downloaded
		mockClient.AssertNumberOfCalls(t, "Do", 1)

		// Verify both images are in memory cache
		cacheMem.mutex.RLock()
		assert.Len(t, cacheMem.cachedImages, 2, "Should have 2 images in memory cache")
		assert.Contains(t, cacheMem.cachedImages, key1, "First key should be in memory cache")
		assert.Contains(t, cacheMem.cachedImages, "key2", "Second key should be in memory cache")
		cacheMem.mutex.RUnlock()
	})

	// Test 3: Prewarming with timeout
	t.Run("Prewarming with timeout", func(t *testing.T) {
		// Skip in normal testing as it takes too long
		t.Skip("Skipping prewarming timeout test - it takes too long")
	})
}

func TestNewImageCache(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &http.Client{}

	// Create a mock config with test port to use temp directory
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "test",
		},
	}

	// Create the cache
	cache, err := NewImageCache(cfg, mockClient)

	// Assertions
	assert.NoError(t, err, "NewImageCache should not return an error")
	assert.NotNil(t, cache, "Cache should not be nil")

	// Test with a non-existent directory by patching os.MkdirAll
	// We need to use a test double approach here
	// Save the original function from the cache package
	originalMkdirAll := osMkdirAll
	defer func() { osMkdirAll = originalMkdirAll }()

	// Mock the MkdirAll function to return an error
	osMkdirAll = func(path string, perm os.FileMode) error {
		return errors.New("mock directory creation error")
	}

	// Try to create the cache again with the mocked function
	badCache, err := NewImageCache(cfg, mockClient)

	// Assertions for the error case
	assert.Error(t, err, "NewImageCache should return an error when directory creation fails")
	assert.Nil(t, badCache, "Cache should be nil when there's an error")
}

func TestGetOrDownloadImage_FileSystemCacheHit(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "image-cache-fs-test")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tmpDir)

	// Create a mock client
	mockClient := new(MockHTTPClient)

	// Create a test image file in the filesystem cache
	testKey := "fs-cache-test"
	testURL := "http://example.com/test.jpg"
	testImageData := []byte("test image data")

	// Create the file in the cache directory
	filePath := filepath.Join(tmpDir, testKey+".jpg")
	err = os.WriteFile(filePath, testImageData, 0644)
	assert.NoError(t, err, "Failed to write test file")

	// Create the cache
	cache := &imageCache{
		cachePath:     tmpDir,
		client:        mockClient,
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
	}
	defer cache.Shutdown()

	// Create test context
	testCtx := setupTestContext()

	// Test getting the image from filesystem cache
	resultPath, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)

	// Assertions
	assert.NoError(t, err, "GetOrDownloadImage should not error")
	assert.Equal(t, filePath, resultPath, "Should return the cached file path")

	// Check that the image was added to memory cache
	cache.mutex.RLock()
	memCachedPath, exists := cache.cachedImages[testKey]
	cache.mutex.RUnlock()

	assert.True(t, exists, "Image should be added to memory cache")
	assert.Equal(t, filePath, memCachedPath, "Cached path should match file path")

	// No HTTP calls should have been made
	mockClient.AssertNotCalled(t, "Do")
}

func TestDownloadImageExtensionHandling(t *testing.T) {
	// Create test environment
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()
	testKey := "content-type-test"

	// Create test cases for different content types
	testCases := []struct {
		name        string
		url         string
		contentType string
		expected    string // expected extension
	}{
		{
			name:        "URL with extension",
			url:         "http://example.com/image.png",
			contentType: "image/png",
			expected:    ".png",
		},
		{
			name:        "No extension, JPEG content type",
			url:         "http://example.com/image",
			contentType: "image/jpeg",
			expected:    ".jpg",
		},
		{
			name:        "No extension, PNG content type",
			url:         "http://example.com/image",
			contentType: "image/png",
			expected:    ".png",
		},
		{
			name:        "No extension, GIF content type",
			url:         "http://example.com/image",
			contentType: "image/gif",
			expected:    ".gif",
		},
		{
			name:        "No extension, WebP content type",
			url:         "http://example.com/image",
			contentType: "image/webp",
			expected:    ".webp",
		},
		{
			name:        "No extension, unknown content type",
			url:         "http://example.com/image",
			contentType: "application/octet-stream",
			expected:    ".jpg", // Default to jpg
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a unique key for each test case
			testKeyWithIndex := fmt.Sprintf("%s-%d", testKey, i)
			testImageData := []byte("test image data")

			// Mock HTTP response
			mockResp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(testImageData)),
				Header:     http.Header{"Content-Type": {tc.contentType}},
			}

			// Set up mock
			mockClient.ExpectedCalls = nil
			mockClient.On("Do", mock.Anything).Return(mockResp, nil)

			// Download the image
			filePath, err := cache.GetOrDownloadImage(testCtx, tc.url, testKeyWithIndex)

			// Assertions
			assert.NoError(t, err, "Download should not fail")
			assert.Contains(t, filePath, tc.expected, "File path should have the correct extension")

			// Verify file exists and has correct content
			fileData, err := os.ReadFile(filePath)
			assert.NoError(t, err, "Should be able to read the file")
			assert.Equal(t, testImageData, fileData, "File content should match test data")
		})
	}
}

func TestCacheMapImages_Errors(t *testing.T) {
	// Create test environment
	cache, tmpDir, mockClient := createTestCache(t)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupTestContext()

	// Test case with download error
	maps := []domain.Map{
		{
			UUID:        "error-map",
			DisplayName: "Error Map",
			Splash:      "http://example.com/error.jpg",
		},
	}

	// Mock HTTP client to return error
	mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("download error"))

	// Try to process maps
	processedMaps, err := cache.CacheMapImages(testCtx, maps)

	// For CacheMapImages, errors during download should be logged but not returned
	assert.NoError(t, err, "CacheMapImages should not return the download error")
	assert.Len(t, processedMaps, 1, "Should still return processed maps")

	// The URL should remain unchanged when download fails
	// This is the actual behavior - CacheMapImages only changes URLs for successful downloads
	// URLs for failed downloads remain unchanged
	assert.Equal(t, "http://example.com/error.jpg", processedMaps[0].Splash,
		"URL should remain unchanged for failed downloads")

	// Verify mock was called
	mockClient.AssertCalled(t, "Do", mock.Anything)
}
