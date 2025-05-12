package api

import (
	"github.com/google/wire"
	"github.com/jungtechou/valomap/api/handler"
	"github.com/jungtechou/valomap/api/handler/cache"
	"github.com/jungtechou/valomap/api/handler/health"
	"github.com/jungtechou/valomap/api/handler/roulette"
	cachesvc "github.com/jungtechou/valomap/service/cache"
)

// HandlerSet contains all API handler providers
var HandlerSet = wire.NewSet(
	health.NewHandler,
	roulette.NewHandler,
	wire.Struct(new(CacheHandlerParams), "*"),
	ProvideCacheHandler,
)

// CacheHandlerParams holds parameters for cache handler creation
type CacheHandlerParams struct {
	ImageCache cachesvc.ImageCache
}

// ProvideCacheHandler creates a new cache handler
func ProvideCacheHandler(params CacheHandlerParams) cache.Handler {
	return cache.NewHandler(params.ImageCache)
}

// NewHandlers provides all API handlers
func NewHandlers(health health.Handler, roulette roulette.Handler, cache cache.Handler) []handler.Handler {
	return []handler.Handler{
		health,
		roulette,
		cache,
	}
}
