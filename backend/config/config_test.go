package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvInt(t *testing.T) {
	// Setup
	os.Clearenv()

	// Test default value
	val := GetEnvInt("TEST_INT", 42)
	assert.Equal(t, 42, val)

	// Test with valid int
	os.Setenv("TEST_INT", "100")
	val = GetEnvInt("TEST_INT", 42)
	assert.Equal(t, 100, val)

	// Test with invalid int
	os.Setenv("TEST_INT", "not-an-int")
	val = GetEnvInt("TEST_INT", 42)
	assert.Equal(t, 42, val) // Should return default

	// Cleanup
	os.Clearenv()
}

func TestGetEnvBool(t *testing.T) {
	// Setup
	os.Clearenv()

	// Test default value
	val := GetEnvBool("TEST_BOOL", true)
	assert.True(t, val)

	// Test with valid bool (true)
	os.Setenv("TEST_BOOL", "true")
	val = GetEnvBool("TEST_BOOL", false)
	assert.True(t, val)

	// Test with valid bool (false)
	os.Setenv("TEST_BOOL", "false")
	val = GetEnvBool("TEST_BOOL", true)
	assert.False(t, val)

	// Test with invalid bool
	os.Setenv("TEST_BOOL", "not-a-bool")
	val = GetEnvBool("TEST_BOOL", true)
	assert.True(t, val) // Should return default

	// Cleanup
	os.Clearenv()
}

func TestGetEnvString(t *testing.T) {
	// Setup
	os.Clearenv()

	// Test default value
	val := GetEnvString("TEST_STRING", "default")
	assert.Equal(t, "default", val)

	// Test with value
	os.Setenv("TEST_STRING", "custom-value")
	val = GetEnvString("TEST_STRING", "default")
	assert.Equal(t, "custom-value", val)

	// Cleanup
	os.Clearenv()
}
