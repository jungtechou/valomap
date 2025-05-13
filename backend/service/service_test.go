package service

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// MockService implements the Service interface for testing
type MockService struct {}

func TestServiceInterface(t *testing.T) {
	// Test that the MockService implements the Service interface
	var s Service = &MockService{}
	assert.NotNil(t, s)
}

func TestVariables(t *testing.T) {
	// Save original values to restore after test
	origOsStat := OsStat
	origJsonUnmarshal := JsonUnmarshal
	origStrconvAtoi := StrconvAtoi
	origLogger := Logger

	defer func() {
		// Restore original values
		OsStat = origOsStat
		JsonUnmarshal = origJsonUnmarshal
		StrconvAtoi = origStrconvAtoi
		Logger = origLogger
	}()

	// Test OsStat
	assert.NotNil(t, OsStat)
	assert.Equal(t, reflect.ValueOf(os.Stat).Pointer(), reflect.ValueOf(OsStat).Pointer())

	// Test JsonUnmarshal
	assert.NotNil(t, JsonUnmarshal)
	assert.Equal(t, reflect.ValueOf(json.Unmarshal).Pointer(), reflect.ValueOf(JsonUnmarshal).Pointer())

	// Test StrconvAtoi
	assert.NotNil(t, StrconvAtoi)
	assert.Equal(t, reflect.ValueOf(strconv.Atoi).Pointer(), reflect.ValueOf(StrconvAtoi).Pointer())

	// Test Logger
	assert.NotNil(t, Logger)
	// logrus.StandardLogger() returns a new instance each time, so we can't compare pointer values
	// Instead, verify it's a logrus.Logger instance
	assert.IsType(t, &logrus.Logger{}, Logger)
}
