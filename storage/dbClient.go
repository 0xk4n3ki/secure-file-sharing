package storage

import (
	"fmt"

	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/models"
)

type dbService struct{}

var DBService = &dbService{}

func (d *dbService) InsertFile(f *models.File, userId string) error {
	query := `INSERT INTO files (filename, user_id, size, s3_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING file_id`

	var file_id string
	err := database.PG_Client.QueryRow(query, f.Filename, userId, f.Size, f.S3Key, f.Created_at, f.Updated_at).Scan(&file_id)
	if err != nil {
		return fmt.Errorf("failed to insert file details in table: %v", err)
	}

	query = `INSERT INTO files_access (file_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err = database.PG_Client.Exec(query, file_id, userId, "owner", f.Created_at)
	if err != nil {
		return fmt.Errorf("faile to insert row int file_access: %v", err)
	}

	return nil
}
