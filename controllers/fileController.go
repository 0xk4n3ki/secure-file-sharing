package controllers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/models"
	"github.com/0xk4n3ki/secure-file-sharing/services"
	"github.com/0xk4n3ki/secure-file-sharing/storage"
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
		fileId := ctx.Param("file_id")

		ownerId, _ := ctx.Get("user_id")
		var perm string
		err := database.PG_Client.QueryRow(`
			SELECT role FROM file_access WHERE user_id=$1 AND file_id=$2`,
			ownerId, fileId,
		).Scan(&perm)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "only owner can share this file"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check file_access"})
			return
		}
		if perm != "owner" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "only owner can share this file"})
			return
		}

		email := ctx.GetHeader("email")
		if email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing collaborator email"})
			return
		}
		var collId string
		err = database.PG_Client.QueryRow(`
			SELECT user_id FROM users WHERE email=$1;`,
			email,
		).Scan(&collId)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the collaborator doesn't exist"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check the existence of collaborator"})
			return
		}
		if collId == ownerId {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot share file with youself"})
			return
		}

		var exists string
		err = database.PG_Client.QueryRow(`
			SELECT user_id FROM file_access WHERE file_id=$1 AND user_id=$2`,
			fileId, collId,
		).Scan(&exists)
		if err == sql.ErrNoRows {

		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "file already shared with this user"})
			return
		}

		var coll_entry models.FileAccess
		err = database.PG_Client.QueryRow(`
			INSERT INTO file_access (file_id, user_id, role, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING file_access_id, file_id, user_id, role, created_at;`,
			fileId, collId, "viewer", time.Now().UTC(),
		).Scan(
			&coll_entry.File_access_id,
			&coll_entry.File_id,
			&coll_entry.User_id,
			&coll_entry.Role,
			&coll_entry.Created_at,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, coll_entry)
	}
}

func Remove() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileId := ctx.Param("file_id")

		ownerId, _ := ctx.Get("user_id")
		var perm string
		err := database.PG_Client.QueryRow(`
			SELECT role FROM file_access WHERE user_id=$1 AND file_id=$2`,
			ownerId, fileId,
		).Scan(&perm)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user doesn't own this file"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check file_access"})
			return
		}
		if perm != "owner" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user doesn't own this file"})
			return
		}

		email := ctx.GetHeader("email")
		if email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing collaborator email"})
			return
		}
		var collId string
		err = database.PG_Client.QueryRow(`
			SELECT user_id FROM users WHERE email=$1;`,
			email,
		).Scan(&collId)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the collaborator doesn't exist"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check the existence of collaborator"})
			return
		}
		if collId == ownerId {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "owner can't remove himself"})
			return
		}

		result, err := database.PG_Client.Exec(`
			DELETE FROM file_access WHERE user_id=$1 AND file_id=$2`,
			collId, fileId,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove collaborator"})
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": "collaborator does not have access to this file"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "collaborator removed successfully"})
	}
}

type fileList struct {
	FileName string `json:"file_name"`
	Role     string `json:"role"`
}

func List() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, _ := ctx.Get("user_id")

		rows, err := database.PG_Client.Query(`
			SELECT f.filename, fa.role
			FROM file_access fa
			JOIN files f ON fa.file_id = f.file_id
			WHERE fa.user_id=$1`, userId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		defer rows.Close()

		data := []fileList{}
		for rows.Next() {
			var tmp fileList
			err := rows.Scan(&tmp.FileName, &tmp.Role)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			data = append(data, tmp)
		}

		if err := rows.Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"files": data})
	}
}

func Download() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, _ := ctx.Get("user_id")
		fileId := ctx.Param("file_id")

		var exist bool
		err := database.PG_Client.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM file_access WHERE user_id=$1 AND file_id=$2)`,
			userId, fileId,
		).Scan(&exist)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exist {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user doesn't have permission"})
			return
		}

		var s3Key, filename string
		err = database.PG_Client.QueryRow(`
			SELECT s3_key, filename FROM files WHERE file_id=$1`,
			fileId).Scan(&s3Key, &filename)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		output, err := storage.S3Service.Download(s3Key)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer output.Body.Close()

		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Length", fmt.Sprintf("%d", *output.ContentLength))

		_, err = io.Copy(ctx.Writer, output.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}

func Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileId := ctx.Param("file_id")
		userId, _ := ctx.Get("user_id")

		var s3key string
		err := database.PG_Client.QueryRow(`
			SELECT s3_key FROM files WHERE owner_id=$1 AND file_id=$2`,
			userId, fileId).Scan(&s3key)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if s3key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user doesn't have access to such file"})
			return
		}

		_, err = database.PG_Client.Exec(`
			DELETE FROM files WHERE owner_id=$1 AND file_id=$2`,
			userId, fileId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = database.PG_Client.Exec(`
			DELETE FROM file_access WHERE file_id=$1`, fileId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = storage.S3Service.Delete(s3key)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "file deleted successfully"})
	}
}
