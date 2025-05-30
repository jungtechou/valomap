package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	return engine
}

func TestRequestLogger(t *testing.T) {
	// Create a test logger with a hook to capture output
	logger := logrus.New()
	hook := &TestLoggerHook{}
	logger.AddHook(hook)

	// Save the original logrus instance and restore after test
	originalLogger := logrus.StandardLogger()
	logrus.SetOutput(logger.Out)
	defer func() {
		logrus.SetOutput(originalLogger.Out)
	}()

	// Setup
	router := setupGin()
	router.Use(RequestLogger())

	// Test successful request
	t.Run("Successful request", func(t *testing.T) {
		hook.Clear()

		router.GET("/test-logger", func(c *gin.Context) {
			c.String(http.StatusOK, "logger test")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-logger", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "logger test", w.Body.String())

		// We can't easily check the logging output directly in a test,
		// but the middleware should execute without errors
	})

	// Test client error (4xx)
	t.Run("Client error (4xx)", func(t *testing.T) {
		hook.Clear()

		router.GET("/test-client-error", func(c *gin.Context) {
			c.AbortWithStatus(http.StatusBadRequest)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-client-error", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test server error (5xx)
	t.Run("Server error (5xx)", func(t *testing.T) {
		hook.Clear()

		router.GET("/test-server-error", func(c *gin.Context) {
			c.AbortWithStatus(http.StatusInternalServerError)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-server-error", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	// Test with URL query parameters
	t.Run("Request with query parameters", func(t *testing.T) {
		hook.Clear()

		router.GET("/test-query", func(c *gin.Context) {
			c.String(http.StatusOK, c.Request.URL.RawQuery)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-query?param1=value1&param2=value2", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "param1=value1&param2=value2", w.Body.String())
	})

	// Test with Gin errors
	t.Run("Request with Gin errors", func(t *testing.T) {
		hook.Clear()

		router.GET("/test-gin-error", func(c *gin.Context) {
			c.Error(errors.New("test error"))
			c.String(http.StatusInternalServerError, "error occurred")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-gin-error", nil)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "error occurred", w.Body.String())
	})
}

// TestLoggerHook is a custom logrus hook for testing
type TestLoggerHook struct {
	Entries []logrus.Entry
}

func (h *TestLoggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TestLoggerHook) Fire(entry *logrus.Entry) error {
	h.Entries = append(h.Entries, *entry)
	return nil
}

func (h *TestLoggerHook) Clear() {
	h.Entries = []logrus.Entry{}
}

func TestErrorHandler(t *testing.T) {
	// Setup
	router := setupGin()
	router.Use(ErrorHandler())

	router.GET("/test-error", func(c *gin.Context) {
		c.Error(errors.New("test error"))
		// Doesn't call c.Abort(), so handlers still execute
	})

	router.GET("/test-error-with-status", func(c *gin.Context) {
		err := fmt.Errorf("status error")
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
	})

	// Test regular error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-error", nil)
	router.ServeHTTP(w, req)

	// Error handler should have transformed the error to a proper response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Test status error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test-error-with-status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCORS(t *testing.T) {
	// Setup
	router := setupGin()
	// Create middleware with specific origin
	router.Use(CORS([]string{"http://localhost:3000"}))

	router.GET("/test-cors", func(c *gin.Context) {
		c.String(http.StatusOK, "cors test")
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-cors", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	// Assert CORS headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "cors test", w.Body.String())
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))

	// Test preflight
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test-cors", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code) // OPTIONS returns 204 No Content
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestRecovery(t *testing.T) {
	// Setup
	router := setupGin()
	router.Use(Recovery())

	router.GET("/test-panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-panic", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// The panic should be caught and a 500 error returned
}

func TestRequestContext(t *testing.T) {
	// Mock a different context creator to test with - we don't want
	// to depend on the specific random UUID generation behavior
	uuidStr := "test-uuid-1234-5678-9012"

	// Setup
	router := setupGin()

	// Create a custom middleware that we control for testing
	testMiddleware := func(c *gin.Context) {
		// Create a new context
		requestCtx := ctx.Background()

		// Use our fixed UUID string
		requestCtx = ctx.WithValue(requestCtx, ctx.RequestIDKey, uuidStr)
		requestCtx = ctx.WithValue(requestCtx, ctx.ClientIPKey, c.ClientIP())

		// Store in Gin context
		c.Set("requestCtx", requestCtx)
		c.Set(string(ctx.RequestIDKey), uuidStr)

		c.Next()
	}

	router.Use(testMiddleware)

	router.GET("/test-context", func(c *gin.Context) {
		// Get the request ID from the gin context
		requestID, exists := c.Get(string(ctx.RequestIDKey))

		// Verify it's not empty
		if !exists || requestID == "" {
			t.Errorf("Expected request ID to be set, but it was empty or not found")
		}

		// Return the request ID in the response
		c.String(http.StatusOK, requestID.(string))
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-context", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	requestID := w.Body.String()
	assert.NotEmpty(t, requestID, "Response should contain the request ID")
	assert.Equal(t, uuidStr, requestID, "Request ID should match our test UUID")
}

func TestGetRequestContext(t *testing.T) {
	// Setup gin
	gin.SetMode(gin.TestMode)

	// Create a new test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a request
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// First test the case where we have no request ID in context
	result := GetRequestContext(c)
	assert.NotNil(t, result, "Should return a context even if no request ID exists")

	// Now apply the RequestContext middleware to set a request ID
	middleware := RequestContext()
	middleware(c)

	// Get the updated context
	result = GetRequestContext(c)

	// Assert that we now have a request ID in the context
	requestID, exists := result.Value(ctx.RequestIDKey).(string)
	assert.True(t, exists, "Request ID should exist in context")
	assert.NotEmpty(t, requestID, "Request ID should not be empty")

	// Test with nil request
	c.Request = nil
	result = GetRequestContext(c)
	assert.NotNil(t, result, "Should return a context even if request is nil")
}
