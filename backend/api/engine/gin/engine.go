package gin

import (
	"github.com/jungtechou/valomap/api/engine"
)

var (
	_ Engine = (*GinEngine)(nil)
)

type Engine interface {
	engine.Engine
}
