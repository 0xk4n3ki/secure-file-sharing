package routes

import (
	"github.com/0xk4n3ki/secure-file-sharing/controllers"
	"github.com/0xk4n3ki/secure-file-sharing/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUser())
}
