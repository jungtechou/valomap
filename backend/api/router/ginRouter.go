package router

import (
	"fmt"
	"strings"

	"github.com/jungtechou/valomap/api/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Ensure GinRouter implements Router interface
var _ Router = (*GinRouter)(nil)

// GinRouter is a Gin implementation of the Router interface
type GinRouter struct {
	routesInfo []handler.RouteInfo
	logger     *logrus.Entry
}

// RegisterAPI registers all routes to the Gin engine
func (r *GinRouter) RegisterAPI(engine *gin.Engine) {
	r.logger.Info("Registering API routes...")

	// Group routes by path prefix for better organization
	routeGroups := make(map[string][]handler.RouteInfo)

	for _, routeInfo := range r.routesInfo {
		// Extract the first part of the path as the group
		parts := strings.Split(strings.Trim(routeInfo.Path, "/"), "/")
		group := "default"
		if len(parts) > 0 {
			group = parts[0]
		}

		routeGroups[group] = append(routeGroups[group], routeInfo)
	}

	// Create API version group
	apiGroup := engine.Group(fmt.Sprintf("/%s", r.PrefixPath()))

	// Register routes by group
	for group, routes := range routeGroups {
		r.logger.WithField("group", group).Infof("Registering %d routes", len(routes))

		for _, routeInfo := range routes {
			path := routeInfo.Path
			// Remove the leading slash if present to avoid double slashes
			if strings.HasPrefix(path, "/") {
				path = path[1:]
			}

			fullPath := fmt.Sprintf("/%s", path)
			r.logger.WithFields(logrus.Fields{
				"method": routeInfo.Method,
				"path":   fullPath,
			}).Debug("Registering route")

			apiGroup.Handle(routeInfo.Method, fullPath, routeInfo.GetFlow()...)
		}
	}

	r.logger.Infof("Registered %d routes successfully", len(r.routesInfo))
}

// GetRoutesInfo returns all route information
func (r *GinRouter) GetRoutesInfo() []handler.RouteInfo {
	return r.routesInfo
}

// ProvideRouteV1 creates a new GinRouter with API v1 prefix
func ProvideRouteV1(receivers ...handler.Handler) *GinRouter {
	logger := logrus.WithField("component", "router")
	routeInfo := extractRouteInfo(receivers...)

	logger.WithField("route_count", len(routeInfo)).Info("Created router with routes")

	return &GinRouter{
		routesInfo: routeInfo,
		logger:     logger,
	}
}

// PrefixPath returns the API version prefix
func (r *GinRouter) PrefixPath() string {
	return "api/v1"
}

// extractRouteInfo extracts route information from handlers
func extractRouteInfo(receivers ...handler.Handler) []handler.RouteInfo {
	var routeInfos []handler.RouteInfo

	if len(receivers) == 0 {
		return routeInfos
	}

	for _, receiver := range receivers {
		routeInfos = append(routeInfos, receiver.GetRouteInfos()...)
	}

	return routeInfos
}
