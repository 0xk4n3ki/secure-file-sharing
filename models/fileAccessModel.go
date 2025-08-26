package models

import "time"

type FileAccess struct {
	ID             int       `json:"id"`
	File_id        string    `json:"file_id"`
	User_id        string    `json:"user_id"`
	Role           string    `json:"role"`
	Created_at     time.Time `json:"created_at"`
	File_access_id string    `json:"file_access_id"`
}
