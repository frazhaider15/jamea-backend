package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jamea/controllers"
)

// SetupRouter configures routes for the service
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	apisV1 := r.Group("/api/v1")
	auth := apisV1.Group("/jamea")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/masool/upload", controllers.UploadMasool)
		auth.POST("/masool-report/upload", controllers.UploadMasoolReport)
		auth.GET("/masool-report", controllers.GetMasoolReport)
		auth.POST("/activities", controllers.PostActivities)
		auth.GET("/activities", controllers.GetActivities)
	}
	return r
}
