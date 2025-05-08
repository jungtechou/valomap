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

	GetRouletteMap(c *gin.Context)
}
