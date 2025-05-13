package cache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock the cache service
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	args := m.Called(ctx, imageURL, cacheKey)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockCacheService) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func (m *MockCacheService) Shutdown() {
	m.Called()
}

// MockImageCache is a mock for the ImageCache interface
type MockImageCache struct {
	mock.Mock
}

func (m *MockImageCache) GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	args := m.Called(ctx, imageURL, cacheKey)
	return args.String(0), args.Error(1)
}

func (m *MockImageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockImageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func (m *MockImageCache) Shutdown() {
	m.Called()
}

func setupTestHandler() (*CacheHandler, string, error) {
	// Create a temporary directory for test cache files
	tempDir, err := os.MkdirTemp("", "cache-handler-test")
	if err != nil {
		return nil, "", err
	}

	// Create a mock cache service
	mockService := &MockCacheService{}

	// Create the handler with the test path
	handler := &CacheHandler{
		imageCachePath: tempDir,
		cacheService:   mockService,
	}

	return handler, tempDir, nil
}

func createTestImage(tempDir, filename string, content []byte) error {
	filePath := filepath.Join(tempDir, filename)
	return os.WriteFile(filePath, content, 0644)
}

func TestGetCachedImage(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create handler and test directory
	handler, tempDir, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test image file
	testImage := []byte("test image data")
	testFilename := "test-image.jpg"
	err = createTestImage(tempDir, testFilename, testImage)
	assert.NoError(t, err, "Failed to create test image")

	// Setup router and test server
	router := gin.New()
	router.GET("/cache/:filename", handler.GetCachedImage)

	// Debug test to see how params are handled
	t.Run("Debug path handling", func(t *testing.T) {
		router := gin.New()
		router.GET("/cache/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			cleanFilename := filepath.Base(filename)
			c.JSON(http.StatusOK, gin.H{
				"original": filename,
				"cleaned":  cleanFilename,
				"equal":    cleanFilename == filename,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/..%2F..%2F..%2Fetc%2Fpasswd", nil)
		router.ServeHTTP(w, req)

		t.Logf("Response: %s", w.Body.String())
	})

	// Test successfully retrieving an image
	t.Run("Successful image retrieval", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/"+testFilename, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testImage, w.Body.Bytes())
		assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=604800", w.Header().Get("Cache-Control"))
	})

	// Test with non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/nonexistent.jpg", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test with missing filename
	t.Run("Missing filename", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test with empty filename
	t.Run("Empty filename", func(t *testing.T) {
		// Create a special route for this test case
		emptyRouter := gin.New()
		emptyRouter.GET("/cache-empty", func(c *gin.Context) {
			// Set filename parameter to empty string directly
			c.Params = []gin.Param{{Key: "filename", Value: ""}}
			handler.GetCachedImage(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache-empty", nil)
		emptyRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Filename is required")
	})

	// Test with path traversal attempt
	t.Run("Path traversal attempt", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/..%2F..%2F..%2Fetc%2Fpasswd", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test explicit path traversal attempt
	t.Run("Explicit path traversal attempt", func(t *testing.T) {
		// Create a special route for this test case
		traversalRouter := gin.New()
		traversalRouter.GET("/cache-traversal", func(c *gin.Context) {
			// Set filename parameter to path traversal directly
			c.Params = []gin.Param{{Key: "filename", Value: "../etc/passwd"}}
			handler.GetCachedImage(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache-traversal", nil)
		traversalRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid filename")
	})

	// Test with If-None-Match header for 304 response
	t.Run("304 Not Modified response", func(t *testing.T) {
		// Get file info to build ETag
		fileInfo, err := os.Stat(filepath.Join(tempDir, testFilename))
		assert.NoError(t, err)
		etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())

		// Make request with matching ETag
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/"+testFilename, nil)
		req.Header.Set("If-None-Match", etag)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotModified, w.Code)
		assert.Empty(t, w.Body.Bytes(), "Response body should be empty for 304")
	})

	// Test with different content types
	t.Run("Different content types", func(t *testing.T) {
		// Create test files with different extensions
		fileTypes := map[string]string{
			"test.png":  "image/png",
			"test.gif":  "image/gif",
			"test.webp": "image/webp",
			"test.bin":  "application/octet-stream",
		}

		for filename, contentType := range fileTypes {
			err = createTestImage(tempDir, filename, testImage)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/cache/"+filename, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, contentType, w.Header().Get("Content-Type"))
		}
	})
}

func TestGetRouteInfos(t *testing.T) {
	handler, tempDir, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}
	defer os.RemoveAll(tempDir)

	routes := handler.GetRouteInfos()
	assert.Len(t, routes, 1, "Should return 1 route")
	assert.Equal(t, http.MethodGet, routes[0].Method)
	assert.Equal(t, "/cache/:filename", routes[0].Path)
}

func TestNewHandler(t *testing.T) {
	// Create a mock cache service
	mockCache := new(MockImageCache)

	// Call the constructor
	handler := NewHandler(mockCache)

	// Assert that the handler is not nil and is of correct type
	assert.NotNil(t, handler)
	assert.Implements(t, (*Handler)(nil), handler)
	assert.Implements(t, (*cache.ImageCache)(nil), mockCache)

	// Type assertion to check the fields
	cacheHandler, ok := handler.(*CacheHandler)
	assert.True(t, ok)
	assert.Equal(t, "/home/appuser/images-cache", cacheHandler.imageCachePath)
	assert.Same(t, mockCache, cacheHandler.cacheService)

	// Verify that we can call GetRouteInfos
	routes := handler.GetRouteInfos()
	assert.Len(t, routes, 1)
	assert.Equal(t, "/cache/:filename", routes[0].Path)
}
