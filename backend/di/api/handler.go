package api

import (
	"github.com/google/wire"
	"github.com/jungtechou/valomap/api/handler"
	"github.com/jungtechou/valomap/api/handler/health"
	"github.com/jungtechou/valomap/api/handler/roulette"
)

// HandlerSet contains all API handler providers
var HandlerSet = wire.NewSet(
	health.NewHandler,
	roulette.NewHandler,
)

// NewHandlers provides all API handlers
func NewHandlers(health health.Handler, roulette roulette.Handler) []handler.Handler {
	return []handler.Handler{
		health,
		roulette,
	}
}
