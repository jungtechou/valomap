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

// GetRouletteMap godoc
// @Summary Get a random map
// @Description Returns a randomly selected Valorant map
// @Tags maps
// @Produce json
// @Success 200 {object} domain.Map
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /map/roulette [get]
func (r *RouletteHandler) GetRouletteMap(c *gin.Context) {
	logger := logrus.WithField("handler", "GetRouletteMap")

	// Create request context
	reqCtx := ctx.Background()

	// Call service to get random map
	rouletteMap, err := r.service.Roulette(reqCtx)
	if err != nil {
		logger.WithError(err).Error("Failed to get random map")

		// Determine appropriate status code and error message
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to retrieve random map"

		if errors.Is(err, roulette.ErrEmptyMapList) {
			statusCode = http.StatusNotFound
			errMsg = "No maps available"
		}

		c.JSON(statusCode, ResponseError{
			Error:   err.Error(),
			Message: errMsg,
			Code:    statusCode,
		})
		return
	}

	logger.WithField("map_name", rouletteMap.DisplayName).Info("Successfully retrieved random map")
	c.JSON(http.StatusOK, rouletteMap)
}

func (r *RouletteHandler) GetRouteInfos() []handler.RouteInfo {
	return []handler.RouteInfo{
		{
			Method:      http.MethodGet,
			Path:        "/map/roulette",
			Middlewares: []gin.HandlerFunc{},
			Handler:     r.GetRouletteMap,
		},
	}
}
