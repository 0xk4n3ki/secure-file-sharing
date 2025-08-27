package routes

import (
	"github.com/0xk4n3ki/secure-file-sharing/controllers"
	"github.com/0xk4n3ki/secure-file-sharing/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	userGroup := incomingRoutes.Group("/users")
	userGroup.Use(middleware.Authenticate())
	{
		userGroup.GET("", controllers.GetUsers())
		userGroup.GET("/:user_id", controllers.GetUser())
	}
}
