package routes

import (
	"github.com/0xk4n3ki/secure-file-sharing/controllers"
	"github.com/0xk4n3ki/secure-file-sharing/middleware"
	"github.com/gin-gonic/gin"
)

func FileRouter(incomingRoutes *gin.Engine) {
	fileGroup := incomingRoutes.Group("/files")
	fileGroup.Use(middleware.Authenticate())
	{
		fileGroup.POST("/upload", controllers.Upload())
		fileGroup.POST("/:id/share", controllers.Share())
		fileGroup.GET("", controllers.List())
		fileGroup.GET("/:id/download", controllers.Download())
		fileGroup.DELETE("/:id", controllers.Delete())
	}
}
