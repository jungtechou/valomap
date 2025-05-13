package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/api/router"
	"github.com/stretchr/testify/assert"
)

// MockEngine implements the Engine interface for testing
type MockEngine struct {
	InitCalled       bool
	StartCalled      bool
	ShutdownCalled   bool
	ServeHTTPCalled  bool
	Router           router.Router
	ShutdownContext  context.Context
	ResponseRecorder *httptest.ResponseRecorder
	Request          *http.Request
}

func (m *MockEngine) Initialize(r router.Router) {
	m.InitCalled = true
	m.Router = r
}

func (m *MockEngine) StartServer() error {
	m.StartCalled = true
	return nil
}

func (m *MockEngine) GracefulShutdown(ctx context.Context) error {
	m.ShutdownCalled = true
	m.ShutdownContext = ctx
	return nil
}

func (m *MockEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.ServeHTTPCalled = true
	m.ResponseRecorder = httptest.NewRecorder()
	m.Request = req
}

func TestEngineInterface(t *testing.T) {
	// Create a mock router
	mockRouter := &MockRouter{}

	// Create our mock engine
	var engine Engine = &MockEngine{}

	// Test the Initialize method
	engine.Initialize(mockRouter)

	// Test the engine can be implemented
	assert.NotNil(t, engine)

	// Type assertion
	mockEngine, ok := engine.(*MockEngine)
	assert.True(t, ok)
	assert.True(t, mockEngine.InitCalled)
	assert.Equal(t, mockRouter, mockEngine.Router)
}

// MockRouter implements router.Router for testing
type MockRouter struct{}

func (r *MockRouter) RegisterAPI(engine *gin.Engine) {
	// Mock implementation that doesn't do anything
}

func (r *MockRouter) PrefixPath() string {
	return "/api/v1"
}
