package router

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/api/handler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Mock handler implementation for testing
type mockHandler struct {
	routes []handler.RouteInfo
}

func (m *mockHandler) GetRouteInfos() []handler.RouteInfo {
	return m.routes
}

func TestProvideRouteV1(t *testing.T) {
	// Setup test handlers
	handler1 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test1",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test1") },
			},
		},
	}

	handler2 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test2",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test2") },
			},
			{
				Method:      http.MethodPost,
				Path:        "/test2",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test2-post") },
			},
		},
	}

	// Create router
	router := ProvideRouteV1(handler1, handler2)

	// Assertions
	assert.NotNil(t, router)
	assert.IsType(t, &GinRouter{}, router)

	// Check if routes were correctly aggregated
	routeInfos := router.GetRoutesInfo()
	assert.Len(t, routeInfos, 3)
}

func TestRegisterAPI(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	handler1 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test1",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test1") },
			},
		},
	}

	// Create router
	router := ProvideRouteV1(handler1)

	// Register routes - this should not panic
	assert.NotPanics(t, func() {
		router.RegisterAPI(engine)
	})

	// Create a new router with different routes to test multiple registrations
	handler2 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test2",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test2") },
			},
		},
	}

	router2 := ProvideRouteV1(handler2)

	// This should not panic either
	assert.NotPanics(t, func() {
		router2.RegisterAPI(engine)
	})
}

func TestPrefixPath(t *testing.T) {
	// Setup
	router := &GinRouter{
		logger: logrus.WithField("test", "test"),
	}

	// Test prefix path
	path := router.PrefixPath()

	// Assertions
	assert.Equal(t, "api/v1", path)
}

func TestExtractRouteInfo(t *testing.T) {
	// Setup
	handler1 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test1",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test1") },
			},
		},
	}

	handler2 := &mockHandler{
		routes: []handler.RouteInfo{
			{
				Method:      http.MethodGet,
				Path:        "/test2",
				Middlewares: []gin.HandlerFunc{},
				Handler:     func(c *gin.Context) { c.String(http.StatusOK, "test2") },
			},
		},
	}

	// Extract route info
	routes := extractRouteInfo(handler1, handler2)

	// Assertions
	assert.Len(t, routes, 2)
	assert.Equal(t, "/test1", routes[0].Path)
	assert.Equal(t, "/test2", routes[1].Path)

	// Test with no handlers
	emptyRoutes := extractRouteInfo()
	assert.Empty(t, emptyRoutes)
}
