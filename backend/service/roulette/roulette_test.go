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
	"github.com/stretchr/testify/require"
)

// MockHttpClient is a mock implementation of the HTTP client
type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockImageCache is a mock implementation of the cache.ImageCache interface
type MockImageCache struct {
	mock.Mock
}

func (m *MockImageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

// Update the method signature to match the interface
func (m *MockImageCache) GetOrDownloadImage(ctx ctx.CTX, url string, fallbackPath string) (string, error) {
	args := m.Called(ctx, url, fallbackPath)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockImageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockImageCache) Shutdown() {
	m.Called()
}

// Helper to create a successful API response
func createSuccessResponse(maps []domain.Map) *http.Response {
	resp := domain.MapResponse{
		Status: 200,
		Data:   maps,
	}
	jsonData, _ := json.Marshal(resp)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(jsonData)),
	}
}

func setupTestContext() ctx.CTX {
	logger := logrus.New()
	logger.Out = io.Discard // Suppress logging during tests
	return ctx.CTX{FieldLogger: logrus.NewEntry(logger)}
}

func TestNewService(t *testing.T) {
	client := &http.Client{}
	mockCache := new(MockImageCache)

	service := NewService(client, mockCache)

	assert.NotNil(t, service, "Service should not be nil")
	assert.Implements(t, (*Service)(nil), service, "Service should implement the Service interface")
}

// Use an HTTP Transport implementation for the tests that satisfies the http.RoundTripper interface
type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestFetchMaps_Success(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	mockCache := new(MockImageCache)
	testCtx := setupTestContext()

	// Test data
	testMaps := []domain.Map{
		{
			UUID:                "map1",
			DisplayName:         "Test Map 1",
			TacticalDescription: "A standard map",
		},
		{
			UUID:                "map2",
			DisplayName:         "Test Map 2",
			TacticalDescription: "Another standard map",
		},
	}

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(createSuccessResponse(testMaps), nil)
	mockCache.On("CacheMapImages", testCtx, testMaps).Return(testMaps, nil)

	// Create service with mocked dependencies
	service := &RouletteService{
		client:     httpClient,
		imageCache: mockCache,
	}

	// Execute
	maps, err := service.fetchMaps(testCtx)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 2, len(maps))
	assert.Equal(t, "map1", maps[0].UUID)
	assert.Equal(t, "map2", maps[1].UUID)

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestFetchMaps_ApiRequestError(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()
	expectedErr := errors.New("network error")

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(nil, expectedErr)

	// Create service with mocked dependencies
	service := &RouletteService{
		client:     httpClient,
		imageCache: nil,
	}

	// Execute
	maps, err := service.fetchMaps(testCtx)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, maps)
	assert.True(t, errors.Is(err, ErrAPIRequest))

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}

func TestFetchMaps_NonOKStatusCode(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()

	// Create a response with non-200 status code
	badResponse := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(badResponse, nil)

	// Create service with mocked dependencies
	service := &RouletteService{
		client:     httpClient,
		imageCache: nil,
	}

	// Execute
	maps, err := service.fetchMaps(testCtx)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, maps)
	assert.True(t, errors.Is(err, ErrAPIResponse))

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}

func TestFetchMaps_InvalidJsonResponse(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()

	// Create a response with invalid JSON
	invalidResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
	}

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(invalidResponse, nil)

	// Create service with mocked dependencies
	service := &RouletteService{
		client:     httpClient,
		imageCache: nil,
	}

	// Execute
	maps, err := service.fetchMaps(testCtx)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, maps)
	assert.True(t, errors.Is(err, ErrAPIResponse))

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}

func TestFetchMaps_EmptyMapList(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()

	// Configure mock behavior with empty map list
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(createSuccessResponse([]domain.Map{}), nil)

	// Create service with mocked dependencies
	service := &RouletteService{
		client:     httpClient,
		imageCache: nil,
	}

	// Execute
	maps, err := service.fetchMaps(testCtx)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, maps)
	assert.Equal(t, ErrEmptyMapList, err)

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}

func TestFilterMaps_BannedMaps(t *testing.T) {
	// Setup
	testCtx := setupTestContext()
	service := &RouletteService{}

	// Test data
	allMaps := []domain.Map{
		{UUID: "map1", DisplayName: "Map 1", TacticalDescription: "Standard"},
		{UUID: "map2", DisplayName: "Map 2", TacticalDescription: "Standard"},
		{UUID: "map3", DisplayName: "Map 3", TacticalDescription: "Standard"},
	}

	filter := MapFilter{
		BannedMapIDs: []string{"map1", "map3"},
	}

	// Execute
	filteredMaps, err := service.filterMaps(testCtx, allMaps, filter)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 1, len(filteredMaps))
	assert.Equal(t, "map2", filteredMaps[0].UUID)
}

func TestFilterMaps_AllMapsBanned(t *testing.T) {
	// Setup
	testCtx := setupTestContext()
	service := &RouletteService{}

	// Test data
	allMaps := []domain.Map{
		{UUID: "map1", DisplayName: "Map 1"},
		{UUID: "map2", DisplayName: "Map 2"},
	}

	filter := MapFilter{
		BannedMapIDs: []string{"map1", "map2"},
	}

	// Execute
	filteredMaps, err := service.filterMaps(testCtx, allMaps, filter)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, filteredMaps)
	assert.Equal(t, ErrNoFilteredMaps, err)
}

func TestFilterMaps_StandardOnly(t *testing.T) {
	// Setup
	testCtx := setupTestContext()
	service := &RouletteService{}

	// Test data
	allMaps := []domain.Map{
		{UUID: "map1", DisplayName: "Map 1", TacticalDescription: "Standard"},
		{UUID: "map2", DisplayName: "Map 2", TacticalDescription: ""},
		{UUID: "map3", DisplayName: "Map 3", TacticalDescription: "Standard"},
	}

	filter := MapFilter{
		StandardOnly: true,
	}

	// Execute
	filteredMaps, err := service.filterMaps(testCtx, allMaps, filter)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 2, len(filteredMaps))
	assert.Equal(t, "map1", filteredMaps[0].UUID)
	assert.Equal(t, "map3", filteredMaps[1].UUID)
}

func TestFilterMaps_NoStandardMaps(t *testing.T) {
	// Setup
	testCtx := setupTestContext()
	service := &RouletteService{}

	// Test data
	allMaps := []domain.Map{
		{UUID: "map1", DisplayName: "Map 1", TacticalDescription: ""},
		{UUID: "map2", DisplayName: "Map 2", TacticalDescription: ""},
	}

	filter := MapFilter{
		StandardOnly: true,
	}

	// Execute
	filteredMaps, err := service.filterMaps(testCtx, allMaps, filter)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, filteredMaps)
	assert.Equal(t, ErrNoStandardMaps, err)
}

func TestGetRandomMap_Success(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()

	// Test data
	testMaps := []domain.Map{
		{
			UUID:                "map1",
			DisplayName:         "Test Map 1",
			TacticalDescription: "A standard map",
		},
		{
			UUID:                "map2",
			DisplayName:         "Test Map 2",
			TacticalDescription: "Another standard map",
		},
	}

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(createSuccessResponse(testMaps), nil)

	// Create a real rand.Rand with fixed seed for deterministic behavior
	source := rand.NewSource(1) // Fixed seed
	rng := rand.New(source)

	service := &RouletteService{
		client: httpClient,
		rng:    rng,
	}

	filter := MapFilter{
		StandardOnly: true,
	}

	// Execute
	randomMap, err := service.GetRandomMap(testCtx, filter)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, randomMap)

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}

func TestGetAllMaps_Success(t *testing.T) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupTestContext()

	// Test data
	testMaps := []domain.Map{
		{
			UUID:                "map1",
			DisplayName:         "Test Map 1",
			TacticalDescription: "A standard map",
		},
		{
			UUID:                "map2",
			DisplayName:         "Test Map 2",
			TacticalDescription: "Another standard map",
		},
	}

	// Configure mock behavior
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(createSuccessResponse(testMaps), nil)

	// Create service with mocked dependencies
	service := &RouletteService{
		client: httpClient,
	}

	// Execute
	maps, err := service.GetAllMaps(testCtx)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 2, len(maps))

	// Verify mocks were called
	mockTransport.AssertExpectations(t)
}
