package helpers

import (
	"log"
	"net/http"
	"time"

	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/models"
	"github.com/gin-gonic/gin"
)

func AddUser(ctx *gin.Context, gUser models.GoogleUser) {
	var count int
	err := database.PG_Client.QueryRow(`
		SELECT COUNT(*)
		FROM users
		where email = $1;`, gUser.Email).Scan(&count)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking user"})
		return
	}
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	var user models.User
	user.First_name = gUser.Given_name
	user.Last_name = gUser.Family_name
	user.Email = gUser.Email
	user.Created_at = time.Now().UTC()
	user.Updated_at = time.Now().UTC()

	var user_id string
	insertErr := database.PG_Client.QueryRow(`
		INSERT INTO users (first_name, last_name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING user_id;
		`, user.First_name,
		user.Last_name,
		user.Email,
		user.Created_at,
		user.Updated_at,
	).Scan(&user_id)
	if insertErr != nil {
		log.Println("error inserting user: ", insertErr)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not created"})
		return
	}

	user.User_id = user_id
	token, refreshToken, _ := GenerateAllTokens(user.Email, user.First_name, user.Last_name, user.User_id)
	user.Token = token
	user.Refresh_token = refreshToken

	_, updateErr := database.PG_Client.Exec(`
		UPDATE users SET token=$1, refresh_token=$2, updated_at=$3 WHERE user_id=$4;
		`, token,
		refreshToken,
		time.Now().UTC(),
		user_id,
	)
	if updateErr != nil {
		log.Println("error updating tokens:", updateErr)
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user_id":       user.User_id,
		"email":         user.Email,
		"token":         user.Token,
		"refresh_token": user.Refresh_token,
	})
}

func LoginUser(ctx *gin.Context, user models.GoogleUser) {
	var foundUser models.User
	err := database.PG_Client.QueryRow(`
		SELECT user_id, first_name, last_name, email, token, refresh_token, created_at, updated_at
		FROM users WHERE email=$1;
		`, user.Email,
	).Scan(
		&foundUser.User_id,
		&foundUser.First_name,
		&foundUser.Last_name,
		&foundUser.Email,
		&foundUser.Token,
		&foundUser.Refresh_token,
		&foundUser.Created_at,
		&foundUser.Updated_at,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User doesn't exist"})
		return
	}

	token, refreshToken, _ := GenerateAllTokens(foundUser.Email, foundUser.First_name, foundUser.Last_name, foundUser.User_id)

	if err := UpdateAllTokens(token, refreshToken, foundUser.User_id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tokens"})
		return
	}

	foundUser.Token = token
	foundUser.Refresh_token = refreshToken

	ctx.JSON(http.StatusOK, gin.H{
		"user_id":       foundUser.User_id,
		"email":         foundUser.Email,
		"first_name":    foundUser.First_name,
		"last_name":     foundUser.Last_name,
		"token":         foundUser.Token,
		"refresh_token": foundUser.Refresh_token,
	})
}
