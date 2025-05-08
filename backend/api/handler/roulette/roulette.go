package roulette

import (
	"errors"
	"net/http"

	"github.com/jungtechou/valomap/api/handler"
	"github.com/jungtechou/valomap/pkg/ctx"
	"github.com/jungtechou/valomap/service/roulette"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ResponseError represents an API error response
type ResponseError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

func NewHandler(service roulette.Service) Handler {
	return &RouletteHandler{service: service}
}

type RouletteHandler struct {
	service roulette.Service
}

// GetMap godoc
// @Summary Get a random map
// @Description Returns a randomly selected Valorant map, with optional filtering
// @Tags maps
// @Produce json
// @Param standard query boolean false "Filter to only standard maps (maps with tactical description)"
// @Param banned query array false "List of map UUIDs to exclude from selection"
// @Success 200 {object} domain.Map
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /map/roulette [get]
func (r *RouletteHandler) GetMap(c *gin.Context) {
	logger := logrus.WithField("handler", "GetMap")

	// Parse the standard query parameter
	standardQuery := c.Query("standard")
	standardOnly := standardQuery == "true" || standardQuery == "1"

	// Parse the banned maps query parameter
	var bannedMapIDs []string
	bannedQuery := c.QueryArray("banned")
	if len(bannedQuery) > 0 {
		bannedMapIDs = bannedQuery
	}

	// Create a filter based on the parameters
	filter := roulette.MapFilter{
		StandardOnly: standardOnly,
		BannedMapIDs: bannedMapIDs,
	}

	// Log the request
	logger.WithFields(logrus.Fields{
		"standard_only": standardOnly,
		"banned_maps":   len(bannedMapIDs),
	}).Info("Processing map roulette request")

	// Create request context
	reqCtx := ctx.Background()

	// Get a random map with the specified filter
	randomMap, err := r.service.GetRandomMap(reqCtx, filter)
	if err != nil {
		r.handleError(c, err, standardOnly)
		return
	}

	// Log successful response
	logger.WithFields(logrus.Fields{
		"map_name":      randomMap.DisplayName,
		"standard_only": standardOnly,
		"banned_maps":   len(bannedMapIDs),
	}).Info("Successfully retrieved random map")

	// Return the result
	c.JSON(http.StatusOK, randomMap)
}

// handleError processes service errors and returns appropriate HTTP responses
func (r *RouletteHandler) handleError(c *gin.Context, err error, standardOnly bool) {
	logger := logrus.WithField("handler", "GetMap")
	logger.WithError(err).Error("Failed to get random map")

	// Default error response
	statusCode := http.StatusInternalServerError
	errMsg := "Failed to retrieve random map"

	// Customize based on the specific error
	switch {
	case errors.Is(err, roulette.ErrEmptyMapList):
		statusCode = http.StatusNotFound
		errMsg = "No maps available"
	case errors.Is(err, roulette.ErrNoStandardMaps):
		statusCode = http.StatusNotFound
		errMsg = "No standard maps available"
	case errors.Is(err, roulette.ErrNoFilteredMaps):
		statusCode = http.StatusNotFound
		errMsg = "All available maps have been banned"
	case errors.Is(err, roulette.ErrAPIRequest), errors.Is(err, roulette.ErrAPIResponse):
		statusCode = http.StatusServiceUnavailable
		errMsg = "Map service unavailable"
	}

	c.JSON(statusCode, ResponseError{
		Error:   err.Error(),
		Message: errMsg,
		Code:    statusCode,
	})
}

// GetAllMaps godoc
// @Summary Get all maps
// @Description Returns a list of all available Valorant maps
// @Tags maps
// @Produce json
// @Success 200 {array} domain.Map
// @Failure 500 {object} ResponseError
// @Router /map/all [get]
func (r *RouletteHandler) GetAllMaps(c *gin.Context) {
	logger := logrus.WithField("handler", "GetAllMaps")
	logger.Info("Processing get all maps request")

	// Create request context
	reqCtx := ctx.Background()

	// Get all maps
	maps, err := r.service.GetAllMaps(reqCtx)
	if err != nil {
		logger.WithError(err).Error("Failed to get all maps")

		// Default error response
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to retrieve maps"

		// Customize based on the specific error
		if errors.Is(err, roulette.ErrAPIRequest) || errors.Is(err, roulette.ErrAPIResponse) {
			statusCode = http.StatusServiceUnavailable
			errMsg = "Map service unavailable"
		}

		c.JSON(statusCode, ResponseError{
			Error:   err.Error(),
			Message: errMsg,
			Code:    statusCode,
		})
		return
	}

	// Log successful response
	logger.WithField("map_count", len(maps)).Info("Successfully retrieved all maps")

	// Return the result
	c.JSON(http.StatusOK, maps)
}

func (r *RouletteHandler) GetRouteInfos() []handler.RouteInfo {
	return []handler.RouteInfo{
		{
			Method:      http.MethodGet,
			Path:        "/map/roulette",
			Middlewares: []gin.HandlerFunc{},
			Handler:     r.GetMap,
		},
		{
			Method:      http.MethodGet,
			Path:        "/map/roulette/standard",
			Middlewares: []gin.HandlerFunc{},
			Handler: func(c *gin.Context) {
				// Force standard mode for the /standard endpoint
				c.Request.URL.RawQuery = "standard=true"
				r.GetMap(c)
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/map/all",
			Middlewares: []gin.HandlerFunc{},
			Handler:     r.GetAllMaps,
		},
	}
}
