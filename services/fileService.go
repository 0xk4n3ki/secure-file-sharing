package services

import (
	"mime/multipart"
	"time"

	"github.com/0xk4n3ki/secure-file-sharing/models"
	"github.com/0xk4n3ki/secure-file-sharing/storage"
	"github.com/google/uuid"
)

type fileService struct{}

var FileService = &fileService{}

func (s *fileService) Upload(userId, filename string, size int64, file multipart.File) (*models.File, error) {
	s3Key := uuid.New().String() + "/" + filename

	err := storage.S3Service.Upload(file, s3Key)
	if err != nil {
		return nil, err
	}

	f := &models.File{
		Filename:   filename,
		Size:       size,
		OwnerID:    userId,
		S3Key:      s3Key,
		Created_at: time.Now().UTC(),
		Updated_at: time.Now().UTC(),
	}

	err = storage.DBService.InsertFile(f, userId)
	if err != nil {
		return nil, err
	}

	return f, nil
}
