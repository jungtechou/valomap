package cache

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockImageCache is a mock for the ImageCache interface
type MockImageCache struct {
	mock.Mock
}

func (m *MockImageCache) GetOrDownloadImage(ctx ctx.CTX, url, key string) (string, error) {
	args := m.Called(ctx, url, key)
	return args.String(0), args.Error(1)
}

func (m *MockImageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func (m *MockImageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockImageCache) Shutdown() {
	m.Called()
}

// Setup test helpers just for these tests
func setupTestPrewarmContext() ctx.CTX {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard) // Silence logs during tests
	return ctx.CTX{
		FieldLogger: logger.WithField("test", true),
	}
}

// mockTransport is a custom RoundTripper for testing
type mockTransport struct {
	mockResponse *http.Response
	mockError    error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.mockResponse, m.mockError
}

func TestNewMapPrewarmer(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	mockClient := &http.Client{}
	apiURL := "http://example.com/api"

	// Create the prewarmer
	prewarmer := NewMapPrewarmer(mockCache, mockClient, apiURL)

	// Assertions
	assert.NotNil(t, prewarmer)
	assert.Equal(t, mockCache, prewarmer.cache)
	assert.Equal(t, mockClient, prewarmer.client)
	assert.Equal(t, apiURL, prewarmer.apiURL)
}

func TestPrewarmMapImagesSuccess(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	apiURL := "http://example.com/api"

	// Mock response data
	responseJSON := `{
		"status": 200,
		"data": [
			{
				"uuid": "map1",
				"displayName": "Test Map 1",
				"splash": "http://example.com/map1/splash.jpg",
				"displayIcon": "http://example.com/map1/icon.png"
			},
			{
				"uuid": "map2",
				"displayName": "Test Map 2",
				"splash": "http://example.com/map2/splash.jpg"
			}
		]
	}`

	// Create mock transport with successful response
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(responseJSON)),
	}

	mockClient := &http.Client{
		Transport: &mockTransport{
			mockResponse: mockResp,
			mockError:    nil,
		},
	}

	// Create the prewarmer with our mocked client
	prewarmer := &MapPrewarmer{
		cache:  mockCache,
		client: mockClient,
		apiURL: apiURL,
	}

	// Expected image URLs to prewarm
	expectedImageURLs := map[string]string{
		"map_map1_splash": "http://example.com/map1/splash.jpg",
		"map_map1_icon":   "http://example.com/map1/icon.png",
		"map_map2_splash": "http://example.com/map2/splash.jpg",
	}

	// Mock cache prewarm
	mockCache.On("PrewarmCache", mock.Anything, expectedImageURLs).Return(nil)

	// Execute the method
	err := prewarmer.PrewarmMapImages()

	// Assertions
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestPrewarmMapImagesHTTPError(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	apiURL := "http://example.com/api"

	// Setup HTTP error
	httpError := errors.New("connection error")
	mockClient := &http.Client{
		Transport: &mockTransport{
			mockResponse: nil,
			mockError:    httpError,
		},
	}

	// Create the prewarmer with our mocked client
	prewarmer := &MapPrewarmer{
		cache:  mockCache,
		client: mockClient,
		apiURL: apiURL,
	}

	// Execute the method
	err := prewarmer.PrewarmMapImages()

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error") // Check for error message containment instead
	mockCache.AssertExpectations(t) // No calls expected
}

func TestPrewarmMapImagesNonOKStatus(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	apiURL := "http://example.com/api"

	// Setup non-OK response
	mockClient := &http.Client{
		Transport: &mockTransport{
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			},
			mockError: nil,
		},
	}

	// Create the prewarmer with our mocked client
	prewarmer := &MapPrewarmer{
		cache:  mockCache,
		client: mockClient,
		apiURL: apiURL,
	}

	// Execute the method
	err := prewarmer.PrewarmMapImages()

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "received non-OK status code: 404")
	mockCache.AssertExpectations(t) // No calls expected
}

func TestPrewarmMapImagesInvalidJSON(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	apiURL := "http://example.com/api"

	// Setup invalid JSON response
	mockClient := &http.Client{
		Transport: &mockTransport{
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString("invalid json")),
			},
			mockError: nil,
		},
	}

	// Create the prewarmer with our mocked client
	prewarmer := &MapPrewarmer{
		cache:  mockCache,
		client: mockClient,
		apiURL: apiURL,
	}

	// Execute the method
	err := prewarmer.PrewarmMapImages()

	// Assertions
	assert.Error(t, err)
	mockCache.AssertExpectations(t) // No calls expected
}

func TestPrewarmMapImagesCacheError(t *testing.T) {
	// Create mocks
	mockCache := new(MockImageCache)
	apiURL := "http://example.com/api"

	// Mock response data
	responseJSON := `{
		"status": 200,
		"data": [
			{
				"uuid": "map1",
				"displayName": "Test Map 1",
				"splash": "http://example.com/map1/splash.jpg"
			}
		]
	}`

	// Setup mock client with good response
	mockClient := &http.Client{
		Transport: &mockTransport{
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(responseJSON)),
			},
			mockError: nil,
		},
	}

	// Create the prewarmer with our mocked client
	prewarmer := &MapPrewarmer{
		cache:  mockCache,
		client: mockClient,
		apiURL: apiURL,
	}

	// Expected image URLs to prewarm
	expectedImageURLs := map[string]string{
		"map_map1_splash": "http://example.com/map1/splash.jpg",
	}

	// Mock cache prewarm with error
	cacheError := errors.New("cache error")
	mockCache.On("PrewarmCache", mock.Anything, expectedImageURLs).Return(cacheError)

	// Execute the method
	err := prewarmer.PrewarmMapImages()

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, cacheError, err)
	mockCache.AssertExpectations(t)
}
