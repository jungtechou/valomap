package middleware

import (
	"fmt"
	"time"

	"github.com/jungtechou/valomap/pkg/ctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestLogger logs each request with request details and response time
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Generate a unique request ID
		requestID := uuid.New().String()
		c.Set(string(ctx.RequestIDKey), requestID)

		// Process request
		c.Next()

		// Calculate response time
		latency := time.Since(start)

		// Get response status
		statusCode := c.Writer.Status()

		// Log details
		if raw != "" {
			path = path + "?" + raw
		}

		entry := logrus.WithFields(logrus.Fields{
			"status":     statusCode,
			"latency":    latency,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"request_id": requestID,
			"user_agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			msg := fmt.Sprintf("%s %s %d %s", c.Request.Method, path, statusCode, latency)
			if statusCode >= 500 {
				entry.Error(msg)
			} else if statusCode >= 400 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

// ErrorHandler handles any errors that occur during request processing
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			// Get request ID from context
			requestID, exists := c.Get(string(ctx.RequestIDKey))
			if !exists {
				requestID = "unknown"
			}

			// Log error
			logrus.WithFields(logrus.Fields{
				"request_id": requestID,
				"errors":     c.Errors,
			}).Error("Request errors")

			// If no response has been sent yet, send a 500 error
			if !c.Writer.Written() {
				c.JSON(500, gin.H{
					"error":      "Internal Server Error",
					"request_id": requestID,
				})
			}
		}
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}

		c.Next()
	}
}

// Recovery middleware for recovering from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID from context
				requestID, exists := c.Get(string(ctx.RequestIDKey))
				if !exists {
					requestID = "unknown"
				}

				// Log error
				logrus.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
				}).Error("Panic recovery")

				// Send response
				c.AbortWithStatusJSON(500, gin.H{
					"error":      "Internal Server Error",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}

// RequestContext adds a context.Context to the Gin context
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new context
		requestCtx := ctx.Background()

		// Generate a request ID and add it to the context
		requestID := uuid.New().String()
		requestCtx = ctx.WithValue(requestCtx, ctx.RequestIDKey, requestID)

		// Add client IP
		requestCtx = ctx.WithValue(requestCtx, ctx.ClientIPKey, c.ClientIP())

		// Store in Gin context
		c.Set("requestCtx", requestCtx)

		// Also set the request ID in the gin context for other middleware to use
		c.Set(string(ctx.RequestIDKey), requestID)

		c.Next()
	}
}

// GetRequestContext retrieves the request context from Gin context
func GetRequestContext(c *gin.Context) ctx.CTX {
	requestCtx, exists := c.Get("requestCtx")
	if !exists {
		// Create a new context if not found
		return ctx.Background()
	}

	return requestCtx.(ctx.CTX)
}
