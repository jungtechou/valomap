package roulette

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"testing"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockTransport implements the RoundTripper interface for testing HTTP clients
type mockTransport struct {
	mock.Mock
}

// RoundTrip implements the RoundTripper interface
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockImageCache is a mock for the ImageCache interface
type MockImageCache struct {
	mock.Mock
}

// CacheMapImages mocks the CacheMapImages method
func (m *MockImageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

// GetOrDownloadImage mocks the GetOrDownloadImage method
func (m *MockImageCache) GetOrDownloadImage(ctx ctx.CTX, url string, id string) (string, error) {
	args := m.Called(ctx, url, id)
	return args.String(0), args.Error(1)
}

// PrewarmCache mocks the PrewarmCache method
func (m *MockImageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

// Shutdown mocks the Shutdown method
func (m *MockImageCache) Shutdown() {
	m.Called()
}

// setupTestContext creates a context for testing
func setupTestContext() ctx.CTX {
	logger := logrus.New()
	logger.Out = io.Discard // Suppress logging during tests
	return ctx.CTX{FieldLogger: logrus.NewEntry(logger)}
}

// TestGetRandomMap tests the basic functionality of GetRandomMap
func TestGetRandomMap(t *testing.T) {
	// Create mock transport and HTTP client
	mt := new(mockTransport)
	httpClient := &http.Client{Transport: mt}

	// Create test maps
	testMaps := []domain.Map{
		{
			UUID:                "map1",
			DisplayName:         "Map One",
			TacticalDescription: "Standard map",
		},
		{
			UUID:                "map2",
			DisplayName:         "Map Two",
			TacticalDescription: "Standard map",
		},
	}

	// Create response with test maps
	resp := domain.MapResponse{
		Status: 200,
		Data:   testMaps,
	}
	jsonData, _ := json.Marshal(resp)

	// Setup HTTP response
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}

	// Configure mock to return the response
	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(httpResp, nil)

	// Create service with deterministic RNG
	service := &RouletteService{
		client: httpClient,
		rng:    rand.New(rand.NewSource(42)), // Use fixed seed for predictable results
	}

	// Get random map
	testCtx := setupTestContext()
	result, err := service.GetRandomMap(testCtx, MapFilter{})

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, []string{"map1", "map2"}, result.UUID)

	// Verify mock expectations
	mt.AssertExpectations(t)
}

// TestFilterMaps tests the map filtering functionality
func TestFilterMaps(t *testing.T) {
	testCtx := setupTestContext()
	service := &RouletteService{
		rng: rand.New(rand.NewSource(42)),
	}

	// Create test maps
	standardMap1 := domain.Map{UUID: "map1", DisplayName: "Map One", TacticalDescription: "Standard map"}
	standardMap2 := domain.Map{UUID: "map2", DisplayName: "Map Two", TacticalDescription: "Standard map"}
	nonStandardMap := domain.Map{UUID: "map3", DisplayName: "Map Three", TacticalDescription: ""}

	testMaps := []domain.Map{standardMap1, standardMap2, nonStandardMap}

	// Test case 1: No filtering
	filteredMaps, err := service.filterMaps(testCtx, testMaps, MapFilter{})
	assert.NoError(t, err)
	assert.Len(t, filteredMaps, 3)

	// Test case 2: Standard maps only
	filteredMaps, err = service.filterMaps(testCtx, testMaps, MapFilter{StandardOnly: true})
	assert.NoError(t, err)
	assert.Len(t, filteredMaps, 2)
	assert.Contains(t, filteredMaps, standardMap1)
	assert.Contains(t, filteredMaps, standardMap2)
	assert.NotContains(t, filteredMaps, nonStandardMap)

	// Test case 3: Banned maps
	filteredMaps, err = service.filterMaps(testCtx, testMaps, MapFilter{BannedMapIDs: []string{"map1"}})
	assert.NoError(t, err)
	assert.Len(t, filteredMaps, 2)
	assert.NotContains(t, filteredMaps, standardMap1)

	// Test case 4: Standard maps only + banned (results in no maps)
	filteredMaps, err = service.filterMaps(testCtx, testMaps,
		MapFilter{StandardOnly: true, BannedMapIDs: []string{"map1", "map2"}})
	assert.Error(t, err)
	assert.Nil(t, filteredMaps)
}

// TestNewService tests the NewService function
func TestNewService(t *testing.T) {
	// Create mock HTTP client and cache
	mockClient := &http.Client{}
	mockCache := new(MockImageCache)

	// Expected mock call
	mockCache.On("PrewarmCache", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Test NewService
	service := NewService(mockClient, mockCache)

	// Verify results
	assert.NotNil(t, service)

	// Assert the service is a *RouletteService
	rouletteService, ok := service.(*RouletteService)
	assert.True(t, ok, "Service should be a *RouletteService")

	// Check fields
	assert.Equal(t, mockClient, rouletteService.client)
	assert.Equal(t, mockCache, rouletteService.imageCache)
	assert.NotNil(t, rouletteService.rng)

	// Verify mock expectations
	mockCache.AssertExpectations(t)
}

// TestGetAllMaps tests the GetAllMaps function
func TestGetAllMaps(t *testing.T) {
	// Create mock transport and HTTP client
	mt := new(mockTransport)
	httpClient := &http.Client{Transport: mt}

	// Create test maps
	testMaps := []domain.Map{
		{
			UUID:        "map1",
			DisplayName: "Map One",
			TacticalDescription: "Standard map",
		},
		{
			UUID:        "map2",
			DisplayName: "Map Two",
			TacticalDescription: "Standard map",
		},
	}

	// Create response with test maps
	resp := domain.MapResponse{
		Status: 200,
		Data:   testMaps,
	}
	jsonData, _ := json.Marshal(resp)

	// Setup HTTP response
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}

	// Configure mock to return the response
	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(httpResp, nil).Once()

	// Create mock cache
	mockCache := new(MockImageCache)
	mockCache.On("CacheMapImages", mock.Anything, mock.Anything).Return(testMaps, nil).Maybe()

	// Create service
	service := &RouletteService{
		client:     httpClient,
		imageCache: mockCache,
		rng:        rand.New(rand.NewSource(42)),
	}

	// Get all maps
	testCtx := setupTestContext()
	maps, err := service.GetAllMaps(testCtx)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 2, len(maps))
	assert.Equal(t, "map1", maps[0].UUID)
	assert.Equal(t, "map2", maps[1].UUID)

	// Verify expectations
	mt.AssertExpectations(t)
	mockCache.AssertExpectations(t)

	// Test error case - HTTP request fails
	reqError := errors.New("request failed")
	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, reqError).Once()

	maps, err = service.GetAllMaps(testCtx)

	// Verify error is returned
	assert.Error(t, err)
	assert.Nil(t, maps)

	// Verify expectations
	mt.AssertExpectations(t)
}

// TestFetchMaps tests the fetchMaps method
func TestFetchMaps(t *testing.T) {
	// Create mock transport and HTTP client
	mt := new(mockTransport)
	httpClient := &http.Client{Transport: mt}

	// Create test maps
	testMaps := []domain.Map{
		{
			UUID:        "map1",
			DisplayName: "Map One",
		},
		{
			UUID:        "map2",
			DisplayName: "Map Two",
		},
	}

	// Create response with test maps
	resp := domain.MapResponse{
		Status: 200,
		Data:   testMaps,
	}
	jsonData, _ := json.Marshal(resp)

	// Setup successful HTTP response
	successResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}

	// Configure mock to return success
	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(successResp, nil).Once()

	// Create service
	service := &RouletteService{
		client: httpClient,
	}

	// Call fetchMaps
	testCtx := setupTestContext()
	maps, err := service.fetchMaps(testCtx)

	// Verify success
	assert.NoError(t, err)
	assert.Equal(t, 2, len(maps))

	// Test with bad JSON response
	badResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
	}

	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(badResp, nil).Once()

	maps, err = service.fetchMaps(testCtx)
	assert.Error(t, err)
	assert.Nil(t, maps)

	// Test with error status code
	errorResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"status": 500, "message": "error"}`))),
	}

	mt.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(errorResp, nil).Once()

	maps, err = service.fetchMaps(testCtx)
	assert.Error(t, err)
	assert.Nil(t, maps)

	// Verify all expectations
	mt.AssertExpectations(t)
}
