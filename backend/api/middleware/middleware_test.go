package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/stretchr/testify/assert"
)

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	return engine
}

func TestRequestLogger(t *testing.T) {
	// Setup
	router := setupGin()
	router.Use(RequestLogger())

	router.GET("/test-logger", func(c *gin.Context) {
		c.String(http.StatusOK, "logger test")
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-logger", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "logger test", w.Body.String())

	// Can't directly test logging output, but we ensure the middleware doesn't break the flow
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
	router.Use(CORS([]string{"*"}))

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
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	// Test preflight
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test-cors", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
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
	// Setup
	router := setupGin()
	router.Use(RequestContext())

	router.GET("/test-context", func(c *gin.Context) {
		// Check that we have a request ID in the context
		reqCtx := c.Request.Context()
		reqID, _ := reqCtx.Value(ctx.RequestIDKey).(string)
		c.String(http.StatusOK, reqID)
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-context", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String()) // Should contain a request ID
}

func TestGetRequestContext(t *testing.T) {
	// Setup test case with valid request context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Create a request with context
	req, _ := http.NewRequest("GET", "/test", nil)
	reqCtx := context.WithValue(context.Background(), ctx.RequestIDKey, "test-id")
	reqWithCtx := req.WithContext(reqCtx)

	c.Request = reqWithCtx

	// Test
	result := GetRequestContext(c)

	// Assert
	assert.NotNil(t, result)
	requestID, _ := result.Value(ctx.RequestIDKey).(string)
	assert.Equal(t, "test-id", requestID)

	// Test with nil request
	c.Request = nil
	result = GetRequestContext(c)
	assert.NotNil(t, result) // Should return a default context
}
