package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWireRegistration(t *testing.T) {
	// Verify InjectorSet is defined (indirect testing of wire.go)
	assert.NotNil(t, InjectorSet, "InjectorSet should not be nil")
}

// Skipping TestBuildInjector because it requires file system permissions
// that aren't available in the test environment.
