package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/0xk4n3ki/secure-file-sharing/config"
	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/routes"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	config.InitGoogleConfig()
	database.RunMigrations()

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": "access granted for api"})
	})

	err = router.Run()
	if err != nil {
		log.Panic("Failed to run the router")
	}
}
