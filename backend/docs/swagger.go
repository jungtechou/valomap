package docs

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Valorant Map Picker API
// @version 1.0
// @description API for selecting random Valorant maps

// @contact.name API Support
// @contact.url https://github.com/jungtechou/valomap
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1
// @schemes http https

// RegisterSwagger adds Swagger documentation endpoints to the router
func RegisterSwagger(router *gin.Engine) {
	// Use the ginSwagger middleware to serve the Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
