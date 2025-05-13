package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCacheHandler(t *testing.T) {
	// Test with empty filename
	t.Run("Empty filename", func(t *testing.T) {
		// Create a special route for this test case
		emptyRouter := gin.New()
		emptyRouter.GET("/cache/*ignored", func(c *gin.Context) {
			// Directly set an empty filename parameter
			c.Params = gin.Params{gin.Param{Key: "filename", Value: ""}}
			handler.GetCachedImage(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/", nil)
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
		// Create a special router that directly sets the params
		traversalRouter := gin.New()
		traversalRouter.GET("/cache/*ignored", func(c *gin.Context) {
			// Set param to simulate path traversal
			c.Params = gin.Params{gin.Param{Key: "filename", Value: "../etc/passwd"}}
			handler.GetCachedImage(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cache/", nil)
		traversalRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid filename")
	})
}
