package cache

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/jungtechou/valomap/api/handler"
	"github.com/jungtechou/valomap/service/cache"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CacheHandler handles requests for cached resources
type CacheHandler struct {
	imageCachePath string
	cacheService   cache.ImageCache
}

// NewHandler creates a new cache handler
func NewHandler(cacheService cache.ImageCache) Handler {
	return &CacheHandler{
		imageCachePath: "./images-cache", // Must match the path in the cache service
		cacheService:   cacheService,
	}
}

// GetCachedImage godoc
// @Summary Get a cached image
// @Description Returns a cached image file
// @Tags cache
// @Produce image/jpeg,image/png,image/gif
// @Param filename path string true "Image filename"
// @Success 200 {file} file "Image file"
// @Failure 404 {object} map[string]string "Image not found"
// @Router /cache/{filename} [get]
func (h *CacheHandler) GetCachedImage(c *gin.Context) {
	logger := logrus.WithField("handler", "GetCachedImage")

	// Get filename from path
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Ensure the filename is safe (no path traversal)
	cleanFilename := filepath.Base(filename)
	if cleanFilename != filename {
		logger.WithFields(logrus.Fields{
			"original": filename,
			"cleaned":  cleanFilename,
		}).Warn("Attempted path traversal in cache request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}

	// Build the full path to the cached file
	filePath := filepath.Join(h.imageCachePath, cleanFilename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.WithField("file", filePath).Info("Requested cache file not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file
	c.File(filePath)
}

// GetRouteInfos implements handler.Handler interface
func (h *CacheHandler) GetRouteInfos() []handler.RouteInfo {
	return []handler.RouteInfo{
		{
			Method:      http.MethodGet,
			Path:        "/cache/:filename",
			Middlewares: []gin.HandlerFunc{},
			Handler:     h.GetCachedImage,
		},
	}
}

// Handler interface defines the methods for cache handlers
type Handler interface {
	handler.Handler
	GetCachedImage(c *gin.Context)
}
