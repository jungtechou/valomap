package api

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/jungtechou/valomap/api/engine/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	// Verify that NewEngine points to the correct function by comparing function names
	assert.NotNil(t, NewEngine, "NewEngine should not be nil")

	// Get function names
	newEngineFuncName := runtime.FuncForPC(reflect.ValueOf(NewEngine).Pointer()).Name()
	ginNewEngineFuncName := runtime.FuncForPC(reflect.ValueOf(gin.NewEngine).Pointer()).Name()

	assert.Equal(t, ginNewEngineFuncName, newEngineFuncName,
		"NewEngine should be an alias for gin.NewEngine")
}
