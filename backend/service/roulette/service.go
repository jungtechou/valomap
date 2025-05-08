package roulette

import (
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service"
)

var (
	_ Service = (*RouletteService)(nil)
)

type Service interface {
	service.Service

	Roulette(ctx ctx.CTX) (*domain.Map, error)
}
