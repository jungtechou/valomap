package roulette

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service/roulette"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock roulette service
type mockRouletteService struct {
	mock.Mock
}

func (m *mockRouletteService) GetRandomMap(ctx ctx.CTX, filter roulette.MapFilter) (*domain.Map, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Map), args.Error(1)
}

func (m *mockRouletteService) GetAllMaps(ctx ctx.CTX) ([]domain.Map, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func setupRouter(handler Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register routes
	routes := handler.GetRouteInfos()
	for _, route := range routes {
		r.Handle(route.Method, route.Path, route.GetFlow()...)
	}

	return r
}

func TestNewHandler(t *testing.T) {
	// Create mock service
	mockService := new(mockRouletteService)

	// Create handler
	handler := NewHandler(mockService)

	// Assert handler was created
	assert.NotNil(t, handler)
	assert.IsType(t, &RouletteHandler{}, handler)
}

func TestGetMap_Success(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)

	// Mock the GetRandomMap call
	testMap := &domain.Map{
		UUID:        "test-map-id",
		DisplayName: "Test Map",
		DisplayIcon: "test-icon-url",
	}

	// Define the expected filter
	expectedFilter := roulette.MapFilter{
		StandardOnly: true,
		BannedMapIDs: []string{"ascent", "bind"},
	}

	// Setup mock to match any context with the expected filter
	mockService.On("GetRandomMap", mock.Anything, expectedFilter).Return(testMap, nil)

	// Create handler and router
	handler := NewHandler(mockService)
	router := setupRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/map/roulette?standard=true&banned=ascent&banned=bind", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response
	assert.Equal(t, "test-map-id", response["uuid"])
	assert.Equal(t, "Test Map", response["displayName"])

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestGetMap_Error(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)

	// Mock the GetRandomMap call to return an error
	mockService.On("GetRandomMap", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))

	// Create handler and router
	handler := NewHandler(mockService)
	router := setupRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/map/roulette", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Parse response
	var response ResponseError
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response contains error
	assert.NotEmpty(t, response.Error)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestGetAllMaps_Success(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)

	// Mock the GetAllMaps call
	testMaps := []domain.Map{
		{UUID: "map1", DisplayName: "Map 1"},
		{UUID: "map2", DisplayName: "Map 2"},
	}
	mockService.On("GetAllMaps", mock.Anything).Return(testMaps, nil)

	// Create handler and router
	handler := NewHandler(mockService)
	router := setupRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/map/all", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response
	assert.Len(t, response, 2)
	assert.Equal(t, "map1", response[0]["uuid"])
	assert.Equal(t, "Map 1", response[0]["displayName"])

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestGetAllMaps_Error(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)

	// Mock the GetAllMaps call to return an error
	mockService.On("GetAllMaps", mock.Anything).Return([]domain.Map{}, errors.New("service error"))

	// Create handler and router
	handler := NewHandler(mockService)
	router := setupRouter(handler)

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/map/all", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestGetRouteInfos(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)
	handler := NewHandler(mockService)

	// Get route infos
	routes := handler.GetRouteInfos()

	// Assertions
	assert.Len(t, routes, 3) // Expect 3 routes: /map/roulette, /map/roulette/standard, and /map/all

	// Verify routes
	assert.Equal(t, http.MethodGet, routes[0].Method)
	assert.Equal(t, "/map/roulette", routes[0].Path)
	assert.NotNil(t, routes[0].Handler)
	assert.Empty(t, routes[0].Middlewares)

	assert.Equal(t, http.MethodGet, routes[1].Method)
	assert.Equal(t, "/map/roulette/standard", routes[1].Path)
	assert.NotNil(t, routes[1].Handler)
	assert.Empty(t, routes[1].Middlewares)

	assert.Equal(t, http.MethodGet, routes[2].Method)
	assert.Equal(t, "/map/all", routes[2].Path)
	assert.NotNil(t, routes[2].Handler)
	assert.Empty(t, routes[2].Middlewares)
}

func TestStandardMapEndpoint(t *testing.T) {
	// Setup
	mockService := new(mockRouletteService)

	// Mock the GetRandomMap call
	testMap := &domain.Map{
		UUID:        "standard-map-id",
		DisplayName: "Standard Map",
		DisplayIcon: "standard-icon-url",
	}

	// Define the expected filter (standard=true is enforced by the endpoint)
	// Use mock.Anything for the BannedMapIDs to avoid nil vs empty slice comparison issues
	mockService.On("GetRandomMap", mock.Anything, mock.MatchedBy(func(filter roulette.MapFilter) bool {
		return filter.StandardOnly == true
	})).Return(testMap, nil)

	// Create handler and router
	handler := NewHandler(mockService)
	router := setupRouter(handler)

	// Create request specifically for the /standard endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/map/roulette/standard", nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response - this should contain the standard map
	assert.Equal(t, "standard-map-id", response["uuid"])
	assert.Equal(t, "Standard Map", response["displayName"])

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestHandleError(t *testing.T) {
	// Setup test cases for different error types
	testCases := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrEmptyMapList",
			err:            roulette.ErrEmptyMapList,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "No maps available",
		},
		{
			name:           "ErrNoStandardMaps",
			err:            roulette.ErrNoStandardMaps,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "No standard maps available",
		},
		{
			name:           "ErrNoFilteredMaps",
			err:            roulette.ErrNoFilteredMaps,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "All available maps have been banned",
		},
		{
			name:           "ErrAPIRequest",
			err:            roulette.ErrAPIRequest,
			expectedStatus: http.StatusServiceUnavailable,
			expectedMsg:    "Map service unavailable",
		},
		{
			name:           "ErrAPIResponse",
			err:            roulette.ErrAPIResponse,
			expectedStatus: http.StatusServiceUnavailable,
			expectedMsg:    "Map service unavailable",
		},
		{
			name:           "Generic error",
			err:            errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to retrieve random map",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := new(mockRouletteService)
			handler := &RouletteHandler{service: mockService}

			// Create a new Gin context
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Request object
			req, _ := http.NewRequest("GET", "/test", nil)
			c.Request = req

			// Call handleError
			handler.handleError(c, tc.err, false)

			// Assertions
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Parse response
			var response ResponseError
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Verify response contains expected error info
			assert.Equal(t, tc.err.Error(), response.Error)
			assert.Equal(t, tc.expectedMsg, response.Message)
			assert.Equal(t, tc.expectedStatus, response.Code)
		})
	}
}
