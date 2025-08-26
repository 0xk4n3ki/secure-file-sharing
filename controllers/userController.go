package controllers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/0xk4n3ki/secure-file-sharing/config"
	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/helpers"
	"github.com/0xk4n3ki/secure-file-sharing/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate = validator.New()

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := config.AppConfig.GoogleLoginConfig.AuthCodeURL("login")
		ctx.Redirect(http.StatusSeeOther, url)
	}
}

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := config.AppConfig.GoogleLoginConfig.AuthCodeURL("signup")
		ctx.Redirect(http.StatusSeeOther, url)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage <= 0 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		var totalCount int
		err = database.PG_Client.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalCount)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error counting users"})
			return
		}

		rows, err := database.PG_Client.Query(`
			SELECT user_id, first_name, last_name, email, token, refresh_token, created_at, updated_at
			FROM users
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, recordPerPage, startIndex)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching users"})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			if err := rows.Scan(
				&u.User_id,
				&u.First_name,
				&u.Last_name,
				&u.Email,
				&u.Token,
				&u.Refresh_token,
				&u.Created_at,
				&u.Updated_at,
			); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error scanning users"})
				return
			}
			users = append(users, u)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"total_count": totalCount,
			"users":       users,
		})
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user models.User
		userId := ctx.Param("user_id")
		if _, err := uuid.Parse(userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
			return
		}

		err := database.PG_Client.QueryRow(`
			SELECT user_id, first_name, last_name, email, token, refresh_token, created_at, updated_at
			FROM users WHERE user_id=$1
		`, userId).Scan(
			&user.User_id,
			&user.First_name,
			&user.Last_name,
			&user.Email,
			&user.Token,
			&user.Refresh_token,
			&user.Created_at,
			&user.Updated_at,
		)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error fetching user",
				"details": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func GoogleCallback() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		state := ctx.Query("state")
		code := ctx.Query("code")

		token, err := config.AppConfig.GoogleLoginConfig.Exchange(ctx.Request.Context(), code)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Code-Token exchange failed"})
			return
		}

		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
			return
		}
		defer resp.Body.Close()

		userData, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Json Parsing failed"})
			return
		}

		var user models.GoogleUser
		if err := json.Unmarshal(userData, &user); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
			return
		}

		err = Validate.Struct(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "missing data returned by google"})
			return
		}

		switch state {
		case "signup":
			helpers.AddUser(ctx, user)
		case "login":
			helpers.LoginUser(ctx, user)
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid State"})
		}
	}
}

func RefreshToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshToken := ctx.GetHeader("refresh_token")
		if refreshToken == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
			return
		}

		claims, msg := helpers.ValidateToken(refreshToken)
		if msg != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		newToken, newRefreshToken, _ := helpers.GenerateAllTokens(
			claims.Email,
			claims.First_name,
			claims.Last_name,
			claims.Uid,
		)
		if err := helpers.UpdateAllTokens(newToken, newRefreshToken, claims.Uid); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tokens"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"token":         newToken,
			"refresh_token": newRefreshToken,
		})
	}
}
