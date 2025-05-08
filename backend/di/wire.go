//go:build wireinject
// +build wireinject

package di

import (
	"github.com/jungtechou/valomap/di/api"
	"github.com/jungtechou/valomap/di/service"

	"github.com/google/wire"
)

func BuildInjector() (*Injector, func(), error) {
	wire.Build(
		// Configuration
		ProvideConfig,

		// API Components
		api.NewEngine,
		api.RouteSet,
		api.HandlerSet,
		api.NewHandlers,

		// Services
		service.ServiceSet,

		// Injector
		InjectorSet,
	)
	return new(Injector), nil, nil
}
