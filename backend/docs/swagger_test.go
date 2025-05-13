package docs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterSwagger(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register swagger
	RegisterSwagger(router)

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test that the swagger endpoint is registered
	resp, err := http.Get(server.URL + "/swagger/index.html")
	assert.NoError(t, err)

	// Should return a 200 OK or a 301 Moved Permanently
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMovedPermanently,
		"Expected status 200 OK or 301 Moved Permanently, got %d", resp.StatusCode)
}
