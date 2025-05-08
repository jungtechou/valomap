package roulette

import (
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service"
)

// MapFilter defines the filtering options for map selection
type MapFilter struct {
	StandardOnly bool
}

var (
	_ Service = (*RouletteService)(nil)
)

type Service interface {
	service.Service

	// GetRandomMap returns a random map filtered by the provided options
	GetRandomMap(ctx ctx.CTX, filter MapFilter) (*domain.Map, error)
}
