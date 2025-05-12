package cache

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
		imageCachePath: "/home/appuser/images-cache", // Must match the path in the cache service
		cacheService:   cacheService,
	}
}

// GetCachedImage godoc
// @Summary Get a cached image
// @Description Returns a cached image file by filename, serving content directly from the cache directory
// @Tags cache
// @Accept json
// @Produce image/jpeg,image/png,image/gif
// @Param filename path string true "Image filename" example:"map_split.jpg"
// @Success 200 {file} file "The requested image file"
// @Failure 400 {object} map[string]string "Bad request - invalid filename"
// @Failure 404 {object} map[string]string "Image not found in cache"
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
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		logger.WithField("file", filePath).Info("Requested cache file not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Get file extension to determine content type
	ext := filepath.Ext(cleanFilename)
	var contentType string
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	default:
		contentType = "application/octet-stream"
	}

	// Set cache control headers
	c.Header("Cache-Control", "public, max-age=604800") // 7 days
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("ETag", fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size()))

	// Check If-None-Match header for 304 responses
	if match := c.GetHeader("If-None-Match"); match != "" {
		etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
		if match == etag {
			c.Status(http.StatusNotModified)
			return
		}
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
