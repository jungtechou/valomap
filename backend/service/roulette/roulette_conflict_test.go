package roulette

import (
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/stretchr/testify/assert"
)

// mockRouletteService is a testable variant of RouletteService with replaceable functionality
type mockRouletteService struct {
	RouletteService
	mockFetchMaps func(ctx ctx.CTX) ([]domain.Map, error)
}

// Override fetchMaps to use the mock if set
func (m *mockRouletteService) fetchMaps(ctx ctx.CTX) ([]domain.Map, error) {
	if m.mockFetchMaps != nil {
		return m.mockFetchMaps(ctx)
	}
	return m.RouletteService.fetchMaps(ctx)
}

// TestConcurrentRandomMapAccess tests that concurrent access to GetRandomMap doesn't cause conflicts
func TestConcurrentRandomMapAccess(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Setup a real HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Use a shared RNG to test for potential thread safety issues
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Create mock service
	service := &mockRouletteService{
		RouletteService: RouletteService{
			client: client,
			rng:    rng,
		},
	}

	// Set mock implementation
	service.mockFetchMaps = func(ctx ctx.CTX) ([]domain.Map, error) {
		// Generate random maps
		maps := make([]domain.Map, 10)
		for i := 0; i < 10; i++ {
			maps[i] = domain.Map{
				UUID:                "map-" + string(rune(i+65)),
				DisplayName:         "Test Map " + string(rune(i+65)),
				TacticalDescription: "Standard map",
			}
		}
		// Simulate some delay
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		return maps, nil
	}

	// Run concurrent requests
	const numGoroutines = 10
	const requestsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Track any issues with concurrent access
	var errLock sync.Mutex
	var errors []error

	// Run concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a context for this goroutine
			testCtx := setupTestContext()

			// Make multiple requests per goroutine with different filters
			for j := 0; j < requestsPerGoroutine; j++ {
				// Create a different filter for each request
				filter := MapFilter{
					StandardOnly: j%2 == 0,
					BannedMapIDs: []string{"map-A", "map-C"},
				}

				// Get a random map
				result, err := service.GetRandomMap(testCtx, filter)

				// Check for errors
				if err != nil {
					errLock.Lock()
					errors = append(errors, err)
					errLock.Unlock()
				} else if result == nil {
					errLock.Lock()
					errors = append(errors, assert.AnError)
					errLock.Unlock()
				}

				// Small sleep to increase chance of race conditions
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)))
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	assert.Empty(t, errors, "Concurrent access should not cause errors")
}

// TestConcurrentFilterMapAccess tests that concurrent access to filterMaps doesn't cause conflicts
func TestConcurrentFilterMapAccess(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Use a shared RNG
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Create service
	service := &RouletteService{
		rng: rng,
	}

	// Generate test maps
	testMaps := createTestMaps(20)

	// Create different filters
	filters := []MapFilter{
		{}, // No filter
		{StandardOnly: true},
		{BannedMapIDs: []string{"map-A", "map-C", "map-E"}},
		{StandardOnly: true, BannedMapIDs: []string{"map-A", "map-C"}},
	}

	// Run concurrent requests
	const numGoroutines = 5
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Track results and errors
	var errLock sync.Mutex
	var errors []error

	// Run concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a context for this goroutine
			testCtx := setupTestContext()

			// Make multiple filter requests per goroutine
			for j := 0; j < requestsPerGoroutine; j++ {
				// Use a different filter for each request
				filter := filters[j%len(filters)]

				// Apply the filter
				_, err := service.filterMaps(testCtx, testMaps, filter)

				// Check for errors
				if err != nil {
					// Only store unexpected errors
					if !(filter.StandardOnly && err == ErrNoStandardMaps) {
						errLock.Lock()
						errors = append(errors, err)
						errLock.Unlock()
					}
				}

				// Small sleep to increase chance of race conditions
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)))
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	assert.Empty(t, errors, "Concurrent access should not cause errors")
}

// TestRaceConditions tests for potential race conditions when accessing the shared random number generator
func TestRaceConditions(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Create mock service with shared RNG
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Add HTTP client to prevent nil pointer dereference
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	service := &mockRouletteService{
		RouletteService: RouletteService{
			rng:    rng,
			client: client, // Initialize the client
		},
	}

	// Set mock implementation
	service.mockFetchMaps = func(ctx ctx.CTX) ([]domain.Map, error) {
		return createTestMaps(10), nil
	}

	// Define a function to run in multiple goroutines
	runRandomMapSelection := func(wg *sync.WaitGroup, filter MapFilter) {
		defer wg.Done()
		ctx := setupTestContext()

		for i := 0; i < 10; i++ {
			_, _ = service.GetRandomMap(ctx, filter)
			time.Sleep(time.Millisecond)
		}
	}

	// Create multiple goroutines to access the service concurrently
	var wg sync.WaitGroup
	const numGoroutines = 5
	wg.Add(numGoroutines)

	// Start goroutines with different filters
	for i := 0; i < numGoroutines; i++ {
		filter := MapFilter{
			StandardOnly: i%2 == 0,
		}
		go runRandomMapSelection(&wg, filter)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// If we get here without deadlocks or panics, the test passes
	// Note: To detect actual race conditions, run with -race flag:
	// go test -race ./...
}
