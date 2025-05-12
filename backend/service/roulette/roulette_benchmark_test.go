package roulette

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// setupBenchmarkContext creates a context for benchmarking with no logging output
func setupBenchmarkContext() ctx.CTX {
	logger := logrus.New()
	logger.Out = io.Discard // Suppress logging during benchmarks
	return ctx.CTX{FieldLogger: logrus.NewEntry(logger)}
}

// createTestMaps generates a number of test maps
func createTestMaps(count int) []domain.Map {
	maps := make([]domain.Map, count)
	for i := 0; i < count; i++ {
		tacticalDesc := "Standard map"
		if i%5 == 0 {
			tacticalDesc = "" // Make some non-standard maps
		}

		maps[i] = domain.Map{
			UUID:                 "map-" + string(rune(i+65)),
			DisplayName:          "Test Map " + string(rune(i+65)),
			TacticalDescription:  tacticalDesc,
			DisplayIcon:          "https://example.com/icon" + string(rune(i+65)),
			Splash:               "https://example.com/splash" + string(rune(i+65)),
			ListViewIcon:         "https://example.com/listview" + string(rune(i+65)),
			ListViewIconTall:     "https://example.com/listviewtall" + string(rune(i+65)),
		}
	}
	return maps
}

// BenchmarkFilterMaps benchmarks the performance of map filtering
func BenchmarkFilterMaps(b *testing.B) {
	testCtx := setupBenchmarkContext()
	service := &RouletteService{
		rng: rand.New(rand.NewSource(42)), // Fixed seed for reproducibility
	}

	testCases := []struct {
		name      string
		mapCount  int
		filter    MapFilter
		setupFunc func([]domain.Map) MapFilter
	}{
		{
			name:     "NoFiltering_10Maps",
			mapCount: 10,
			filter:   MapFilter{},
		},
		{
			name:     "NoFiltering_100Maps",
			mapCount: 100,
			filter:   MapFilter{},
		},
		{
			name:     "StandardOnly_10Maps",
			mapCount: 10,
			filter:   MapFilter{StandardOnly: true},
		},
		{
			name:     "StandardOnly_100Maps",
			mapCount: 100,
			filter:   MapFilter{StandardOnly: true},
		},
		{
			name:     "BannedMaps_Half_10Maps",
			mapCount: 10,
			setupFunc: func(maps []domain.Map) MapFilter {
				banned := make([]string, len(maps)/2)
				for i := 0; i < len(maps)/2; i++ {
					banned[i] = maps[i].UUID
				}
				return MapFilter{BannedMapIDs: banned}
			},
		},
		{
			name:     "BannedMaps_Half_100Maps",
			mapCount: 100,
			setupFunc: func(maps []domain.Map) MapFilter {
				banned := make([]string, len(maps)/2)
				for i := 0; i < len(maps)/2; i++ {
					banned[i] = maps[i].UUID
				}
				return MapFilter{BannedMapIDs: banned}
			},
		},
		{
			name:     "ComplexFilter_100Maps",
			mapCount: 100,
			setupFunc: func(maps []domain.Map) MapFilter {
				banned := make([]string, len(maps)/4)
				for i := 0; i < len(maps)/4; i++ {
					banned[i] = maps[i*2].UUID
				}
				return MapFilter{StandardOnly: true, BannedMapIDs: banned}
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Generate test maps
			testMaps := createTestMaps(tc.mapCount)

			// Setup filter if needed
			filter := tc.filter
			if tc.setupFunc != nil {
				filter = tc.setupFunc(testMaps)
			}

			// Reset the timer before the benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = service.filterMaps(testCtx, testMaps, filter)
			}
		})
	}
}

// BenchmarkGetRandomMap benchmarks the performance of getting a random map
func BenchmarkGetRandomMap(b *testing.B) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	testCtx := setupBenchmarkContext()

	// Create test data with different sizes
	smallMapSet := createTestMaps(5)
	mediumMapSet := createTestMaps(20)
	largeMapSet := createTestMaps(50)

	testCases := []struct {
		name    string
		maps    []domain.Map
		filter  MapFilter
	}{
		{
			name:   "SmallMapSet_NoFilter",
			maps:   smallMapSet,
			filter: MapFilter{},
		},
		{
			name:   "MediumMapSet_NoFilter",
			maps:   mediumMapSet,
			filter: MapFilter{},
		},
		{
			name:   "LargeMapSet_NoFilter",
			maps:   largeMapSet,
			filter: MapFilter{},
		},
		{
			name:   "SmallMapSet_StandardOnly",
			maps:   smallMapSet,
			filter: MapFilter{StandardOnly: true},
		},
		{
			name:   "MediumMapSet_StandardOnly",
			maps:   mediumMapSet,
			filter: MapFilter{StandardOnly: true},
		},
		{
			name:   "LargeMapSet_StandardOnly",
			maps:   largeMapSet,
			filter: MapFilter{StandardOnly: true},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create mock response
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBuffer(createAPIResponse(tc.maps))),
			}

			// Configure mock behavior
			mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(resp, nil).Maybe()

			// Create service with deterministic RNG
			source := rand.NewSource(time.Now().UnixNano())
			service := &RouletteService{
				client: httpClient,
				rng:    rand.New(source),
			}

			// Reset timer before benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Need to reset the response body between iterations
				resp.Body = io.NopCloser(bytes.NewBuffer(createAPIResponse(tc.maps)))
				_, _ = service.GetRandomMap(testCtx, tc.filter)
			}

			// Stop timer and cleanup
			b.StopTimer()
			mockTransport.AssertExpectations(b)
		})
	}
}

// Helper to create API response JSON
func createAPIResponse(maps []domain.Map) []byte {
	resp := domain.MapResponse{
		Status: 200,
		Data:   maps,
	}
	jsonData, _ := json.Marshal(resp)
	return jsonData
}

// BenchmarkFetchMaps benchmarks the map fetching functionality
func BenchmarkFetchMaps(b *testing.B) {
	// Setup
	mockTransport := new(mockTransport)
	httpClient := &http.Client{Transport: mockTransport}
	mockCache := new(MockImageCache)
	testCtx := setupBenchmarkContext()

	// Create test data
	testMaps := createTestMaps(20)

	// Configure mock behavior for successful response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBuffer(createAPIResponse(testMaps))),
	}
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(resp, nil).Maybe()

	// Mock the cache behavior
	mockCache.On("CacheMapImages", testCtx, mock.AnythingOfType("[]domain.Map")).Return(testMaps, nil).Maybe()

	testCases := []struct {
		name       string
		withCache  bool
		mapCount   int
	}{
		{
			name:      "WithoutCache_SmallResponse",
			withCache: false,
			mapCount:  5,
		},
		{
			name:      "WithCache_SmallResponse",
			withCache: true,
			mapCount:  5,
		},
		{
			name:      "WithoutCache_MediumResponse",
			withCache: false,
			mapCount:  20,
		},
		{
			name:      "WithCache_MediumResponse",
			withCache: true,
			mapCount:  20,
		},
		{
			name:      "WithoutCache_LargeResponse",
			withCache: false,
			mapCount:  50,
		},
		{
			name:      "WithCache_LargeResponse",
			withCache: true,
			mapCount:  50,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create test maps
			benchMaps := createTestMaps(tc.mapCount)

			// Create service with or without cache
			var service *RouletteService
			if tc.withCache {
				service = &RouletteService{
					client:     httpClient,
					imageCache: mockCache,
				}
			} else {
				service = &RouletteService{
					client: httpClient,
				}
			}

			// Reset timer before benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Reset the response body between iterations
				resp.Body = io.NopCloser(bytes.NewBuffer(createAPIResponse(benchMaps)))
				_, _ = service.fetchMaps(testCtx)
			}

			// Stop timer and cleanup
			b.StopTimer()
		})
	}
}
