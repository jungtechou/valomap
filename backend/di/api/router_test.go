package api

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/jungtechou/valomap/api/router"
	"github.com/stretchr/testify/assert"
)

func TestRouteSet(t *testing.T) {
	// Verify the RouteSet is properly defined
	assert.NotNil(t, RouteSet, "RouteSet should not be nil")
}

func TestGinRouterSet(t *testing.T) {
	// Verify the GinRouterSet is properly defined
	assert.NotNil(t, GinRouterSet, "GinRouterSet should not be nil")
}

func TestProvideRouteV1(t *testing.T) {
	// Test that ProvideRouteV1 is an alias for router.ProvideRouteV1 by comparing function names
	assert.NotNil(t, ProvideRouteV1, "ProvideRouteV1 should not be nil")
	assert.NotNil(t, router.ProvideRouteV1, "router.ProvideRouteV1 should not be nil")

	// Get function names
	ourFuncName := runtime.FuncForPC(reflect.ValueOf(ProvideRouteV1).Pointer()).Name()
	routerFuncName := runtime.FuncForPC(reflect.ValueOf(router.ProvideRouteV1).Pointer()).Name()

	assert.Equal(t, routerFuncName, ourFuncName,
		"ProvideRouteV1 should be an alias for router.ProvideRouteV1")
}
