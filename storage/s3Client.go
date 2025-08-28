package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/0xk4n3ki/secure-file-sharing/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Service struct{}

var S3Service = &s3Service{}

func (s *s3Service) Upload(file multipart.File, s3key string) error {
	c := context.Background()

	_, err := config.S3Client.PutObject(c, &s3.PutObjectInput{
		Bucket:               aws.String(config.BucketName),
		Key:                  aws.String(s3key),
		Body:                 file,
		ServerSideEncryption: "aws:kms",
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %w", s3key, err)
	}
	return nil
}

func (s *s3Service) Download(s3key string) (*s3.GetObjectOutput, error) {
	c := context.Background()

	return config.S3Client.GetObject(c, &s3.GetObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(s3key),
	})
}
