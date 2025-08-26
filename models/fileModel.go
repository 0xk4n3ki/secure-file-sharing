package models

import "time"

type File struct {
	ID           int       `json:"id"`
	Filename     string    `json:"filename"`
	Filepath     string    `json:"filepath"`
	OwnerID      string    `json:"owner_id"`
	Size         int64     `json:"size"`
	S3Key        string    `json:"s3_key"`
	Created_at   time.Time `json:"created_at"`
	Updated_at   time.Time `json:"updated_at"`
	EncryptedDEK []byte    `json:"-"`
	File_id      string    `json:"file_id"`
}
