package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(handler Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register routes
	routeInfos := handler.GetRouteInfos()
	for _, info := range routeInfos {
		r.Handle(info.Method, info.Path, info.GetFlow()...)
	}

	return r
}

func TestNewHandler(t *testing.T) {
	// Create handler
	handler := NewHandler()

	// Assertions
	assert.NotNil(t, handler)
	assert.IsType(t, &HealthHandler{}, handler)
}

func TestHealthCheck(t *testing.T) {
	// Setup
	handler := NewHandler()
	router := setupTestRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.NotZero(t, response.Timestamp)
	assert.NotZero(t, response.Memory)
}

func TestPing(t *testing.T) {
	// Setup
	handler := NewHandler()
	router := setupTestRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "pong", response["message"])
}

func TestGetRouteInfos(t *testing.T) {
	// Setup
	handler := NewHandler()

	// Call function
	routes := handler.GetRouteInfos()

	// Assertions
	assert.Len(t, routes, 2)

	// Check health route
	healthRoute := routes[0]
	assert.Equal(t, http.MethodGet, healthRoute.Method)
	assert.Equal(t, "/health", healthRoute.Path)
	assert.NotNil(t, healthRoute.Handler)

	// Check ping route
	pingRoute := routes[1]
	assert.Equal(t, http.MethodGet, pingRoute.Method)
	assert.Equal(t, "/ping", pingRoute.Path)
	assert.NotNil(t, pingRoute.Handler)
}
