package cache

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// Benchmark the GetCachedImage handler
func BenchmarkGetCachedImage(b *testing.B) {
	// Setup Gin in release mode for benchmarks
	gin.SetMode(gin.ReleaseMode)

	// Create handler and test directory
	handler, tempDir, err := setupTestHandler()
	if err != nil {
		b.Fatalf("Failed to setup test handler: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test image file
	testImage := make([]byte, 100*1024) // 100KB file
	for i := range testImage {
		testImage[i] = byte(i % 256)
	}
	testFilename := "benchmark-image.jpg"
	err = createTestImage(tempDir, testFilename, testImage)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	// Setup router
	router := gin.New()
	router.GET("/cache/:filename", handler.GetCachedImage)

	// Create HTTP request
	req, _ := http.NewRequest("GET", "/cache/"+testFilename, nil)

	// Setup recorder
	w := httptest.NewRecorder()

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// Benchmark the GetCachedImage handler with ETag (304 response)
func BenchmarkGetCachedImageWithETag(b *testing.B) {
	// Setup Gin in release mode for benchmarks
	gin.SetMode(gin.ReleaseMode)

	// Create handler and test directory
	handler, tempDir, err := setupTestHandler()
	if err != nil {
		b.Fatalf("Failed to setup test handler: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test image file
	testImage := make([]byte, 100*1024) // 100KB file
	for i := range testImage {
		testImage[i] = byte(i % 256)
	}
	testFilename := "benchmark-etag-image.jpg"
	err = createTestImage(tempDir, testFilename, testImage)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	// Get file info to build ETag
	fileInfo, err := os.Stat(filepath.Join(tempDir, testFilename))
	if err != nil {
		b.Fatalf("Failed to get file info: %v", err)
	}
	etag := "\"" + fileInfo.ModTime().String() + "\""

	// Setup router
	router := gin.New()
	router.GET("/cache/:filename", handler.GetCachedImage)

	// Create HTTP request with ETag
	req, _ := http.NewRequest("GET", "/cache/"+testFilename, nil)
	req.Header.Set("If-None-Match", etag)

	// Setup recorder
	w := httptest.NewRecorder()

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// Benchmark different image sizes
func BenchmarkImageSizes(b *testing.B) {
	// Image sizes to test
	sizes := []struct {
		name string
		size int
	}{
		{"Small", 10 * 1024},   // 10KB
		{"Medium", 100 * 1024}, // 100KB
		{"Large", 1024 * 1024}, // 1MB
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			// Setup Gin in release mode for benchmarks
			gin.SetMode(gin.ReleaseMode)

			// Create handler and test directory
			handler, tempDir, err := setupTestHandler()
			if err != nil {
				b.Fatalf("Failed to setup test handler: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test image file
			testImage := make([]byte, size.size)
			for i := range testImage {
				testImage[i] = byte(i % 256)
			}
			testFilename := "benchmark-" + size.name + "-image.jpg"
			err = createTestImage(tempDir, testFilename, testImage)
			if err != nil {
				b.Fatalf("Failed to create test image: %v", err)
			}

			// Setup router
			router := gin.New()
			router.GET("/cache/:filename", handler.GetCachedImage)

			// Create HTTP request
			req, _ := http.NewRequest("GET", "/cache/"+testFilename, nil)

			// Setup recorder
			w := httptest.NewRecorder()

			// Run benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				router.ServeHTTP(w, req)
			}
		})
	}
}
