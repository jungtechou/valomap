package cache

import (
	"io"
	"os"
	"sync"
	"testing"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// Setup benchmark-specific context
func setupBenchContext() ctx.CTX {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Silence logs during benchmarks
	return ctx.CTX{
		FieldLogger: logger.WithField("benchmark", true),
	}
}

// Create a benchmark-specific version of createTestCache
func createBenchTestCache(b *testing.B) (ImageCache, string, *MockHTTPClient) {
	// Create temporary directory for test cache
	tmpDir, err := os.MkdirTemp("", "image-cache-benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create mock client
	mockClient := new(MockHTTPClient)

	// Create cache with test directory and mock client
	cache := &imageCache{
		cachePath:     tmpDir,
		client:        mockClient,
		cachedImages:  make(map[string]string),
		downloadQueue: make(chan downloadTask, 10),
		mutex:         sync.RWMutex{},
	}

	// Start worker goroutines
	for range 2 {
		go cache.downloadWorker()
	}

	return cache, tmpDir, mockClient
}

func BenchmarkGetOrDownloadImage_CachedInMemory(b *testing.B) {
	cache, tmpDir, mockClient := createBenchTestCache(b)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupBenchContext()
	testURL := "http://example.com/test.jpg"
	testKey := "test-key"
	testImageData := []byte("test image data")

	// Mock HTTP client response
	mockResp := createMockImageResponse(testImageData, "image/jpeg")
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// First download to populate the cache
	_, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)
	if err != nil {
		b.Fatalf("Failed to prime cache: %v", err)
	}

	// Benchmark memory cache retrieval
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetOrDownloadImage(testCtx, testURL, testKey)
	}
}

func BenchmarkGetOrDownloadImage_CachedOnDisk(b *testing.B) {
	cache, tmpDir, mockClient := createBenchTestCache(b)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupBenchContext()
	testURL := "http://example.com/test.jpg"
	testKey := "test-key"
	testImageData := []byte("test image data")

	// Mock HTTP client response
	mockResp := createMockImageResponse(testImageData, "image/jpeg")
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// First download to populate the cache
	_, err := cache.GetOrDownloadImage(testCtx, testURL, testKey)
	if err != nil {
		b.Fatalf("Failed to prime cache: %v", err)
	}

	// Clear memory cache but keep file on disk
	imageCacheImpl := cache.(*imageCache)
	imageCacheImpl.mutex.Lock()
	imageCacheImpl.cachedImages = make(map[string]string)
	imageCacheImpl.mutex.Unlock()

	// Benchmark disk cache retrieval
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetOrDownloadImage(testCtx, testURL, testKey)
	}
}

func BenchmarkGetOrDownloadImage_Download(b *testing.B) {
	cache, tmpDir, mockClient := createBenchTestCache(b)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupBenchContext()
	testImageData := []byte("test image data")

	// Mock HTTP client response
	mockResp := createMockImageResponse(testImageData, "image/jpeg")
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Benchmark fresh downloads
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testURL := "http://example.com/test.jpg"
		testKey := "test-key-" + string(rune(i))
		cache.GetOrDownloadImage(testCtx, testURL, testKey)
	}
}

func BenchmarkCacheMapImages(b *testing.B) {
	cache, tmpDir, mockClient := createBenchTestCache(b)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupBenchContext()
	testImageData := []byte("test image data")

	// Create 10 maps with images
	maps := make([]domain.Map, 10)
	for i := 0; i < 10; i++ {
		maps[i] = domain.Map{
			UUID:        "map" + string(rune(i)),
			DisplayName: "Test Map " + string(rune(i)),
			Splash:      "http://example.com/splash" + string(rune(i)) + ".jpg",
			DisplayIcon: "http://example.com/icon" + string(rune(i)) + ".png",
		}

		// Setup mocks for each URL
		mockResp := createMockImageResponse(testImageData, "image/jpeg")
		mockClient.On("Do", mock.Anything).Return(mockResp, nil)
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.CacheMapImages(testCtx, maps)
	}
}

func BenchmarkPrewarmCache(b *testing.B) {
	cache, tmpDir, mockClient := createBenchTestCache(b)
	defer os.RemoveAll(tmpDir)
	defer cache.Shutdown()

	testCtx := setupBenchContext()
	testImageData := []byte("test image data")

	// Create URL map with 20 images
	urlMap := make(map[string]string)
	for i := 0; i < 20; i++ {
		key := "key" + string(rune(i))
		url := "http://example.com/img" + string(rune(i)) + ".jpg"
		urlMap[key] = url

		// Setup mocks for each URL
		mockResp := createMockImageResponse(testImageData, "image/jpeg")
		mockClient.On("Do", mock.Anything).Return(mockResp, nil)
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.PrewarmCache(testCtx, urlMap)
	}
}
