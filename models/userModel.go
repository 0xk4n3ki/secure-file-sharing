package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id"`
	First_name    string    `json:"first_name"`
	Last_name     string    `json:"last_name"`
	Email         string    `json:"email"`
	Token         string    `json:"token,omitempty"`
	Refresh_token string    `json:"refresh_token,omitempty"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
	User_id       string    `json:"user_id"`
}
