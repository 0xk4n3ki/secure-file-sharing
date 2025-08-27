package services

import (
	"mime/multipart"
	"time"

	"github.com/0xk4n3ki/secure-file-sharing/models"
	"github.com/0xk4n3ki/secure-file-sharing/storage"
)

type fileService struct{}

var FileService = &fileService{}

func (s *fileService) Upload(userId, filename string, size int64, file multipart.File) (*models.File, error) {
	finalFileName, err := storage.DBService.GetAvailableFileName(userId, filename)
	if err != nil {
		return nil, err
	}

	s3Key := userId + "/" + finalFileName

	err = storage.S3Service.Upload(file, s3Key)
	if err != nil {
		return nil, err
	}

	f := &models.File{
		Filename:   finalFileName,
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
