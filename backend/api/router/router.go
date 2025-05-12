package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jungtechou/valomap/api/handler/cache"
)

type Router interface {
	PrefixPath() string
	RegisterAPI(engine *gin.Engine)
}

type Handlers struct {
	Cache cache.Handler
}
