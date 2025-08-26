package routes

import (
	"github.com/0xk4n3ki/secure-file-sharing/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("users/signup", controllers.Signup())
	incomingRoutes.GET("users/login", controllers.Login())

	incomingRoutes.GET("/google_callback", controllers.GoogleCallback())
	incomingRoutes.POST("/users/refresh", controllers.RefreshToken())
}
