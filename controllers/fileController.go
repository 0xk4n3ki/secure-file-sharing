package controllers

import (
	"net/http"

	"github.com/0xk4n3ki/secure-file-sharing/services"
	"github.com/gin-gonic/gin"
)

func Upload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.GetString("user_id")
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
			return
		}
		defer file.Close()

		uploadedFile, err := services.FileService.Upload(userId, header.Filename, header.Size, file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, uploadedFile)
	}
}

func Share() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func List() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func Download() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
