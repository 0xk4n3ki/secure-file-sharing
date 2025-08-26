package routes

import (
	"github.com/0xk4n3ki/secure-file-sharing/controllers"
	"github.com/gin-gonic/gin"
)

func FileRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/files/upload", controllers.Upload())
	incomingRoutes.POST("/files/:id/share", controllers.Share())
	incomingRoutes.GET("/files", controllers.List())
	incomingRoutes.GET("/files/:id/download", controllers.Download())
	incomingRoutes.DELETE("/files/:id", controllers.Delete())
}
