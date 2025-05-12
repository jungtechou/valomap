package api

import (
	"testing"

	"github.com/jungtechou/valomap/api/handler/cache"
	"github.com/jungtechou/valomap/api/handler/health"
	"github.com/jungtechou/valomap/api/handler/roulette"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockImageCache is a mock implementation of cache.ImageCache
type MockImageCache struct {
	mock.Mock
}

func (m *MockImageCache) GetOrDownloadImage(ctx ctx.CTX, imageURL, cacheKey string) (string, error) {
	args := m.Called(ctx, imageURL, cacheKey)
	return args.String(0), args.Error(1)
}

func (m *MockImageCache) PrewarmCache(ctx ctx.CTX, urlMap map[string]string) error {
	args := m.Called(ctx, urlMap)
	return args.Error(0)
}

func (m *MockImageCache) CacheMapImages(ctx ctx.CTX, maps []domain.Map) ([]domain.Map, error) {
	args := m.Called(ctx, maps)
	return args.Get(0).([]domain.Map), args.Error(1)
}

func (m *MockImageCache) Shutdown() {
	m.Called()
}

func TestProvideCacheHandler(t *testing.T) {
	// Create a mock image cache
	mockCache := &MockImageCache{}

	// Create params with the mock cache
	params := CacheHandlerParams{
		ImageCache: mockCache,
	}

	// Call the function under test
	handler := ProvideCacheHandler(params)

	// Verify the handler is created correctly
	assert.NotNil(t, handler, "Handler should not be nil")
	assert.IsType(t, &cache.CacheHandler{}, handler, "Handler should be of type *cache.CacheHandler")
}

func TestNewHandlers(t *testing.T) {
	// Create mock handlers
	healthHandler := &health.HealthHandler{}
	rouletteHandler := &roulette.RouletteHandler{}
	cacheHandler := &cache.CacheHandler{}

	// Call the function under test
	handlers := NewHandlers(healthHandler, rouletteHandler, cacheHandler)

	// Verify the handlers are returned correctly
	assert.Len(t, handlers, 3, "Should return 3 handlers")
	assert.Contains(t, handlers, healthHandler, "Should contain health handler")
	assert.Contains(t, handlers, rouletteHandler, "Should contain roulette handler")
	assert.Contains(t, handlers, cacheHandler, "Should contain cache handler")
}
