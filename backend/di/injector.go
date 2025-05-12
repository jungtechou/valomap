package di

import (
	"net/http"

	"github.com/jungtechou/valomap/api/engine/gin"
	"github.com/jungtechou/valomap/config"
	"github.com/jungtechou/valomap/service/cache"

	"github.com/google/wire"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

// Injector holds all application dependencies
type Injector struct {
	HttpEngine *gin.GinEngine
	Config     *config.Config
	ImageCache cache.ImageCache
	HTTPClient *http.Client
}

// ProvideConfig provides the application configuration
func ProvideConfig() (*config.Config, error) {
	return config.Load()
}
