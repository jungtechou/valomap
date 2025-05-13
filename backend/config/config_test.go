package config

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func TestSetDefaults(t *testing.T) {
	// Create a new viper instance
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Verify defaults are set correctly
	assert.Equal(t, "3000", v.GetString("server.port"))
	assert.Equal(t, 10*time.Second, v.GetDuration("server.read_timeout"))
	assert.Equal(t, 10*time.Second, v.GetDuration("server.write_timeout"))
	assert.Equal(t, 5*time.Second, v.GetDuration("server.shutdown_timeout"))

	assert.Equal(t, "info", v.GetString("logging.level"))
	assert.Equal(t, "text", v.GetString("logging.format"))
	assert.Equal(t, false, v.GetBool("logging.report_caller"))

	assert.Equal(t, "/api", v.GetString("api.base_path"))
	assert.Equal(t, "v1", v.GetString("api.version"))
	assert.Equal(t, "https://valorant-api.com/v1/maps", v.GetString("api.map_api_url"))
	assert.Equal(t, 5*time.Second, v.GetDuration("api.request_timeout"))

	assert.Equal(t, false, v.GetBool("redis.enabled"))
	assert.Equal(t, "localhost:6379", v.GetString("redis.address"))
	assert.Equal(t, "", v.GetString("redis.password"))
	assert.Equal(t, 0, v.GetInt("redis.db"))

	assert.Equal(t, []string{"*"}, v.GetStringSlice("security.allowed_origins"))
	assert.Contains(t, v.GetStringSlice("security.allowed_methods"), "GET")
	assert.Equal(t, int64(1024*1024*8), v.GetInt64("security.max_body_size"))
}

func TestSetupLogger(t *testing.T) {
	// Test 1: JSON format
	config := LoggingConfig{
		Level:        "debug",
		Format:       "json",
		ReportCaller: true,
	}
	setupLogger(config)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	_, ok := logrus.StandardLogger().Formatter.(*logrus.JSONFormatter)
	assert.True(t, ok)
	assert.True(t, logrus.StandardLogger().ReportCaller)

	// Test 2: Text format with invalid level (should default to info)
	config = LoggingConfig{
		Level:        "invalid_level",
		Format:       "text",
		ReportCaller: false,
	}
	setupLogger(config)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	_, ok = logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
	assert.True(t, ok)
	assert.False(t, logrus.StandardLogger().ReportCaller)
}

func TestLoad(t *testing.T) {
	// Save original environment variables
	originalPort := os.Getenv("VALOMAP_SERVER_PORT")
	originalLevel := os.Getenv("VALOMAP_LOGGING_LEVEL")
	originalMapURL := os.Getenv("VALOMAP_API_MAP_API_URL")

	defer func() {
		// Restore original environment variables
		setEnvOrUnset("VALOMAP_SERVER_PORT", originalPort)
		setEnvOrUnset("VALOMAP_LOGGING_LEVEL", originalLevel)
		setEnvOrUnset("VALOMAP_API_MAP_API_URL", originalMapURL)
	}()

	// Test default config (unset env vars first)
	os.Unsetenv("VALOMAP_SERVER_PORT")
	os.Unsetenv("VALOMAP_LOGGING_LEVEL")
	os.Unsetenv("VALOMAP_API_MAP_API_URL")

	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "3000", config.Server.Port)
	assert.Equal(t, "info", config.Logging.Level)
}

// Helper function to set or unset an environment variable
func setEnvOrUnset(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func TestGetEnvFunctions(t *testing.T) {
	// Save original environment variables
	originalTestInt := os.Getenv("TEST_INT")
	originalTestBool := os.Getenv("TEST_BOOL")
	originalTestStr := os.Getenv("TEST_STR")

	defer func() {
		// Restore original environment variables
		setEnvOrUnset("TEST_INT", originalTestInt)
		setEnvOrUnset("TEST_BOOL", originalTestBool)
		setEnvOrUnset("TEST_STR", originalTestStr)
	}()

	// Test GetEnvInt
	os.Unsetenv("TEST_INT")
	assert.Equal(t, 42, GetEnvInt("TEST_INT", 42)) // Default when not set
	os.Setenv("TEST_INT", "100")
	assert.Equal(t, 100, GetEnvInt("TEST_INT", 42)) // Parsed value
	os.Setenv("TEST_INT", "invalid")
	assert.Equal(t, 42, GetEnvInt("TEST_INT", 42)) // Default when invalid

	// Test GetEnvBool
	os.Unsetenv("TEST_BOOL")
	assert.Equal(t, true, GetEnvBool("TEST_BOOL", true)) // Default when not set
	os.Setenv("TEST_BOOL", "false")
	assert.Equal(t, false, GetEnvBool("TEST_BOOL", true)) // Parsed value
	os.Setenv("TEST_BOOL", "invalid")
	assert.Equal(t, true, GetEnvBool("TEST_BOOL", true)) // Default when invalid

	// Test GetEnvString
	os.Unsetenv("TEST_STR")
	assert.Equal(t, "default", GetEnvString("TEST_STR", "default")) // Default when not set
	os.Setenv("TEST_STR", "value")
	assert.Equal(t, "value", GetEnvString("TEST_STR", "default")) // Actual value
}

// Helper function to split environment variable key-value pairs
func splitKeyValue(env string) (string, string, bool) {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return env[:i], env[i+1:], true
		}
	}
	return env, "", false
}
