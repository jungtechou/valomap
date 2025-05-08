package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds all configuration for the server
type Config struct {
	Server   ServerConfig
	Logging  LoggingConfig
	API      APIConfig
	Redis    RedisConfig
	Security SecurityConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// LoggingConfig holds all logging-related configuration
type LoggingConfig struct {
	Level        string
	Format       string
	ReportCaller bool
}

// APIConfig holds all API-related configuration
type APIConfig struct {
	BasePath       string
	Version        string
	MapAPIURL      string
	RequestTimeout time.Duration
}

// RedisConfig holds all Redis-related configuration
type RedisConfig struct {
	Enabled  bool
	Address  string
	Password string
	DB       int
}

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	MaxBodySize    int64
}

// Load loads the configuration from environment variables, files, and defaults
func Load() (*Config, error) {
	v := viper.New()

	// Set default configurations
	setDefaults(v)

	// Load config file if it exists
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.valorant-map-picker")
	v.AddConfigPath("/etc/valorant-map-picker")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		logrus.Info("No config file found, using default values and environment variables")
	}

	// Load environment variables
	v.SetEnvPrefix("VALOMAP")
	v.AutomaticEnv()

	// Create config
	config := &Config{
		Server: ServerConfig{
			Port:            v.GetString("server.port"),
			ReadTimeout:     v.GetDuration("server.read_timeout"),
			WriteTimeout:    v.GetDuration("server.write_timeout"),
			ShutdownTimeout: v.GetDuration("server.shutdown_timeout"),
		},
		Logging: LoggingConfig{
			Level:        v.GetString("logging.level"),
			Format:       v.GetString("logging.format"),
			ReportCaller: v.GetBool("logging.report_caller"),
		},
		API: APIConfig{
			BasePath:       v.GetString("api.base_path"),
			Version:        v.GetString("api.version"),
			MapAPIURL:      v.GetString("api.map_api_url"),
			RequestTimeout: v.GetDuration("api.request_timeout"),
		},
		Redis: RedisConfig{
			Enabled:  v.GetBool("redis.enabled"),
			Address:  v.GetString("redis.address"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		Security: SecurityConfig{
			AllowedOrigins: v.GetStringSlice("security.allowed_origins"),
			AllowedMethods: v.GetStringSlice("security.allowed_methods"),
			MaxBodySize:    v.GetInt64("security.max_body_size"),
		},
	}

	setupLogger(config.Logging)

	return config, nil
}

// setDefaults sets the default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "3000")
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 5*time.Second)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("logging.report_caller", false)

	// API defaults
	v.SetDefault("api.base_path", "/api")
	v.SetDefault("api.version", "v1")
	v.SetDefault("api.map_api_url", "https://valorant-api.com/v1/maps")
	v.SetDefault("api.request_timeout", 5*time.Second)

	// Redis defaults
	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.address", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Security defaults
	v.SetDefault("security.allowed_origins", []string{"*"})
	v.SetDefault("security.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("security.max_body_size", 1024*1024*8) // 8MB
}

// setupLogger configures the global logger based on configuration
func setupLogger(config LoggingConfig) {
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set log format
	if config.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set report caller
	logrus.SetReportCaller(config.ReportCaller)
}

// GetEnvInt gets an integer value from an environment variable
func GetEnvInt(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// GetEnvBool gets a boolean value from an environment variable
func GetEnvBool(key string, defaultVal bool) bool {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// GetEnvString gets a string value from an environment variable
func GetEnvString(key, defaultVal string) string {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	return s
}
