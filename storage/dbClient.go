package storage

import (
	"fmt"
	"strings"

	"github.com/0xk4n3ki/secure-file-sharing/database"
	"github.com/0xk4n3ki/secure-file-sharing/models"
)

type dbService struct{}

var DBService = &dbService{}

func (d *dbService) InsertFile(f *models.File, userId string) error {
	query := `INSERT INTO files (filename, owner_id, size, s3_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING file_id`

	var file_id string
	err := database.PG_Client.QueryRow(query, f.Filename, userId, f.Size, f.S3Key, f.Created_at, f.Updated_at).Scan(&file_id)
	if err != nil {
		return fmt.Errorf("failed to insert file details in table: %v", err)
	}

	query = `INSERT INTO file_access (file_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err = database.PG_Client.Exec(query, file_id, userId, "owner", f.Created_at)
	if err != nil {
		return fmt.Errorf("faile to insert row in file_access: %v", err)
	}

	f.File_id = file_id

	return nil
}

func (d *dbService) GetAvailableFileName(userId, filename string) (string, error) {
	var count int
	baseName := filename
	extension := ""

	if dot := strings.LastIndex(filename, "."); dot != -1 {
		baseName = filename[:dot]
		extension = filename[dot:]
	}

	newName := filename
	for i := 1; ; i++ {
		query := `SELECT COUNT(1) FROM files WHERE owner_id=$1 AND filename=$2`
		err := database.PG_Client.QueryRow(query, userId, newName).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("error checking filename conflict: %v", err)
		}

		if count == 0 {
			break
		}
		newName = fmt.Sprintf("%s(%d)%s", baseName, i, extension)
	}

	return newName, nil
}
