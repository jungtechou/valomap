package health

import (
	"github.com/jungtechou/valomap/api/handler"

	"github.com/gin-gonic/gin"
)

// Handler defines the health check handler interface
type Handler interface {
	handler.Handler

	HealthCheck(c *gin.Context)
	Ping(c *gin.Context)
}
