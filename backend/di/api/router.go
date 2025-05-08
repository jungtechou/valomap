package api

import (
	"github.com/jungtechou/valomap/api/router"

	"github.com/google/wire"
)

var RouteSet = wire.NewSet(GinRouterSet)
var GinRouterSet = wire.NewSet(ProvideRouteV1, wire.Bind(new(router.Router), new(*router.GinRouter)))

var ProvideRouteV1 = router.ProvideRouteV1
