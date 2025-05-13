package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockImageCacheForPrewarm is a mock for the ImageCache interface for prewarming tests
type MockImageCacheForPrewarm struct {
	mock.Mock
}

func (m *MockImageCacheForPrewarm) GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	args := m.Called(ctx, imageURL, cacheKey)
	return args.String(0), args.Error(1)
}

func (m *MockImageCacheForPrewarm) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockImageCacheForPrewarm) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func (m *MockImageCacheForPrewarm) Shutdown() {
	m.Called()
}

// MockHTTPTransport is a mock HTTP transport for testing the prewarmer
type MockHTTPTransport struct {
	mock.Mock
}

func (m *MockHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockMapService is a mock for the map service
type MockMapService struct {
	mock.Mock
}

func (m *MockMapService) GetAllMaps(ctx ctx.CTX) ([]domain.Map, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Map), args.Error(1)
}

// TestNewMapPrewarmer tests the NewMapPrewarmer function
func TestNewMapPrewarmer(t *testing.T) {
	// Create mock dependencies
	mockCache := new(MockImageCacheForPrewarm)
	mockClient := &http.Client{}
	apiURL := "https://example.com/api/maps"

	// Test creation of the prewarmer
	prewarmer := NewMapPrewarmer(mockCache, mockClient, apiURL)

	// Verify result
	assert.NotNil(t, prewarmer)
	assert.Equal(t, mockCache, prewarmer.cache)
	assert.Equal(t, mockClient, prewarmer.client)
	assert.Equal(t, apiURL, prewarmer.apiURL)
}

// TestPrewarmMapImages tests the PrewarmMapImages function
func TestPrewarmMapImages(t *testing.T) {
	// Create mock dependencies
	mockCache := new(MockImageCacheForPrewarm)
	mockTransport := new(MockHTTPTransport)

	// Create HTTP client with mock transport
	client := &http.Client{
		Transport: mockTransport,
	}

	// Create test maps
	testMaps := []domain.Map{
		{
			UUID:        "map1",
			DisplayName: "Map One",
			Splash:      "https://example.com/splash1.jpg",
			DisplayIcon: "https://example.com/icon1.jpg",
		},
		{
			UUID:        "map2",
			DisplayName: "Map Two",
			Splash:      "https://example.com/splash2.jpg",
			DisplayIcon: "https://example.com/icon2.jpg",
		},
	}

	// Create API response
	apiResp := domain.MapResponse{
		Status: 200,
		Data:   testMaps,
	}
	jsonData, err := json.Marshal(apiResp)
	require.NoError(t, err)

	// Mock HTTP response
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}

	// Configure mock transport
	apiURL := "https://example.com/api/maps"
	mockTransport.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == apiURL
	})).Return(mockResp, nil)

	// Configure mock cache
	mockCache.On("PrewarmCache", mock.Anything, mock.MatchedBy(func(urlMap map[string]string) bool {
		// Verify the URL map contains the expected entries
		return len(urlMap) == 4
	})).Return(nil)

	// Create prewarmer
	prewarmer := NewMapPrewarmer(mockCache, client, apiURL)

	// Call the function
	err = prewarmer.PrewarmMapImages()

	// Verify results
	require.NoError(t, err)
	mockTransport.AssertExpectations(t)
	mockCache.AssertExpectations(t)

	// Test with HTTP error
	httpErr := new(MockHTTPTransport)
	httpErr.On("RoundTrip", mock.Anything).Return(&http.Response{}, assert.AnError)

	clientErr := &http.Client{Transport: httpErr}
	prewarmerErr := NewMapPrewarmer(mockCache, clientErr, apiURL)

	err = prewarmerErr.PrewarmMapImages()
	assert.Error(t, err)
	httpErr.AssertExpectations(t)

	// Test with non-200 status code
	httpBadStatus := new(MockHTTPTransport)
	badResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"error": "server error"}`))),
	}
	httpBadStatus.On("RoundTrip", mock.Anything).Return(badResp, nil)

	clientBadStatus := &http.Client{Transport: httpBadStatus}
	prewarmerBadStatus := NewMapPrewarmer(mockCache, clientBadStatus, apiURL)

	err = prewarmerBadStatus.PrewarmMapImages()
	assert.Error(t, err)
	httpBadStatus.AssertExpectations(t)
}
