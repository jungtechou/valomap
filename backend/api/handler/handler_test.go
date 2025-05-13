package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRouteInfoGetFlow(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test recorder
	w := httptest.NewRecorder()

	// Create a mock value to track middleware execution
	middlewareCalled1 := false
	middlewareCalled2 := false
	handlerCalled := false

	// Create a test middleware
	middleware1 := func(c *gin.Context) {
		middlewareCalled1 = true
		c.Next()
	}

	middleware2 := func(c *gin.Context) {
		middlewareCalled2 = true
		c.Next()
	}

	// Create a test handler
	handler := func(c *gin.Context) {
		handlerCalled = true
		c.String(http.StatusOK, "test")
	}

	// Create a RouteInfo with middlewares and handler
	routeInfo := &RouteInfo{
		Method:      http.MethodGet,
		Path:        "/test",
		Middlewares: []gin.HandlerFunc{middleware1, middleware2},
		Handler:     handler,
	}

	// Get the flow
	flow := routeInfo.GetFlow()

	// Assertions on the length - we can't directly compare function pointers
	assert.Len(t, flow, 3)

	// Create a context to execute the handlers
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	// Execute each handler in flow
	for _, h := range flow {
		h(c)
	}

	// Verify all functions were called
	assert.True(t, middlewareCalled1, "First middleware should be called")
	assert.True(t, middlewareCalled2, "Second middleware should be called")
	assert.True(t, handlerCalled, "Handler should be called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())

	// Test with no middlewares
	emptyMiddlewaresRoute := &RouteInfo{
		Method:      http.MethodGet,
		Path:        "/test",
		Middlewares: []gin.HandlerFunc{},
		Handler:     handler,
	}

	// Reset recorder and flags
	w = httptest.NewRecorder()
	handlerCalled = false

	emptyFlow := emptyMiddlewaresRoute.GetFlow()
	assert.Len(t, emptyFlow, 1)

	// Execute the handler
	c, _ = gin.CreateTestContext(w)
	req, _ = http.NewRequest("GET", "/test", nil)
	c.Request = req

	emptyFlow[0](c)

	// Verify handler was called
	assert.True(t, handlerCalled, "Handler should be called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())
}
