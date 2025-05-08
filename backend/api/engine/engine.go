package engine

import (
	"context"
	"net/http"

	"github.com/jungtechou/valomap/api/router"
)

type Engine interface {
	Initialize(r router.Router)
	StartServer() error
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	GracefulShutdown(ctx context.Context) error
}
