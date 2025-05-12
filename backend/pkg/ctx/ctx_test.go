package ctx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackground(t *testing.T) {
	// Call Background
	ctx := Background()

	// Assertions
	assert.NotNil(t, ctx)

	// Check that it has the expected fields
	reqID := ctx.Value(RequestIDKey)
	assert.NotNil(t, reqID)
	assert.IsType(t, "", reqID)
}

func TestWithValue(t *testing.T) {
	// Setup
	baseCtx := Background()
	key := ContextKey("testKey")
	value := "testValue"

	// Call WithValue
	resultCtx := WithValue(baseCtx, key, value)

	// Assertions
	assert.NotNil(t, resultCtx)

	// Check that the value was set
	storedValue := resultCtx.Value(key)
	assert.Equal(t, value, storedValue)

	// Check that other values are preserved
	reqID := resultCtx.Value(RequestIDKey)
	assert.NotNil(t, reqID)
}

func TestWithValues(t *testing.T) {
	// Setup
	baseCtx := Background()
	values := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	// Call WithValues
	resultCtx := WithValues(baseCtx, values)

	// Assertions
	assert.NotNil(t, resultCtx)

	// Check that all values were set
	for k, v := range values {
		storedValue := resultCtx.Value(ContextKey(k))
		assert.Equal(t, v, storedValue)
	}

	// Check that other values are preserved
	reqID := resultCtx.Value(RequestIDKey)
	assert.NotNil(t, reqID)
}

func TestWithCancel(t *testing.T) {
	// Setup
	baseCtx := Background()

	// Call WithCancel
	resultCtx, cancel := WithCancel(baseCtx)
	defer cancel()

	// Assertions
	assert.NotNil(t, resultCtx)

	// Check that the cancel function works
	cancel()
	select {
	case <-resultCtx.Done():
		// Context was canceled, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context was not canceled")
	}
}

func TestWithTimeout(t *testing.T) {
	// Setup
	baseCtx := Background()
	timeout := 50 * time.Millisecond

	// Call WithTimeout
	resultCtx, cancel := WithTimeout(baseCtx, timeout)
	defer cancel()

	// Assertions
	assert.NotNil(t, resultCtx)

	// Check that the timeout works
	select {
	case <-resultCtx.Done():
		// Context timed out, which is expected
	case <-time.After(timeout * 2):
		t.Fatal("Context did not time out")
	}
}

func TestRequestIDMethod(t *testing.T) {
	// Setup with custom request ID
	ctx := Background()

	// Test getting request ID
	reqID := ctx.RequestID()
	assert.NotEmpty(t, reqID)
}

func TestElapsedTimeMethod(t *testing.T) {
	// Setup
	ctx := Background()

	// Add a small delay to ensure there's elapsed time
	time.Sleep(5 * time.Millisecond)

	// Test getting elapsed time
	elapsed := ctx.ElapsedTime()

	// Assertions
	assert.True(t, elapsed >= 5*time.Millisecond)
}

func TestLogElapsedMethod(t *testing.T) {
	// Setup
	ctx := Background()

	// Call LogElapsed (we can't easily check the log output, but we can check it doesn't panic)
	assert.NotPanics(t, func() {
		ctx.LogElapsed("test-operation")
	})
}

func TestFromContext(t *testing.T) {
	// Setup with various contexts
	stdCtx := context.Background()
	customCtx := Background()

	// Test converting from standard context
	result1 := FromContext(stdCtx)
	assert.NotNil(t, result1)
	assert.IsType(t, CTX{}, result1)

	// Test converting from custom context
	result2 := FromContext(customCtx)
	assert.NotNil(t, result2)
	assert.IsType(t, CTX{}, result2)
}
