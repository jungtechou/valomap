package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapPool_AvailableMaps(t *testing.T) {
	// Setup pool with test maps
	pool := &MapPool{
		Maps: []Map{
			{UUID: "map1", DisplayName: "Map 1"},
			{UUID: "map2", DisplayName: "Map 2"},
			{UUID: "map3", DisplayName: "Map 3"},
		},
	}

	// Test with no banned maps
	available := pool.AvailableMaps([]string{})
	assert.Len(t, available, 3)

	// Test with one banned map
	available = pool.AvailableMaps([]string{"map2"})
	assert.Len(t, available, 2)
	// Check specific maps are included/excluded
	found := false
	for _, m := range available {
		if m.UUID == "map2" {
			found = true
			break
		}
	}
	assert.False(t, found, "Banned map should not be present")

	// Test with all maps banned
	available = pool.AvailableMaps([]string{"map1", "map2", "map3"})
	assert.Len(t, available, 0)
}
