package services

import (
	"bytes"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StorageService struct {
	client     *s3.S3
	bucketName string
	region     string
	endpoint   string
}

func NewStorageService() *StorageService {
	accessKey := "DO004NAPPW8KZY2LKPMJ"
	secretKey := "ed/lOnhZIruavG2UqVtkO+wZ3KqbjlgxQpGEyigoLw4"
	region := "sgp1"
	endpoint := "sgp1.digitaloceanspaces.com" // <-- ONLY the region endpoint
	bucket := "ust-translation"

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	return &StorageService{
		client:     s3.New(sess),
		bucketName: bucket,
		region:     region,
		endpoint:   fmt.Sprintf("https://%s.%s", bucket, endpoint), // for public URL
	}
}

func (s *StorageService) UploadFile(file multipart.File, fileName string) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", err
	}

	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s", s.endpoint, fileName)
	return url, nil
}

func (s *StorageService) DeleteFile(fileName string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	})
	return err
}
