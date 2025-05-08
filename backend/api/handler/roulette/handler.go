package roulette

import (
	"github.com/jungtechou/valomap/api/handler"

	"github.com/gin-gonic/gin"
)

var (
	_ Handler = (*RouletteHandler)(nil)
)

type Handler interface {
	handler.Handler

	// GetMap returns a random map, optionally filtered to standard maps only
	GetMap(c *gin.Context)
}
