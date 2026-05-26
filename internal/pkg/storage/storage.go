package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var S3Client *s3.Client

const S3BaseURL = "https://%s.s3.%s.amazonaws.com/%s"

func InitS3() error {
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		return errors.New("AWS_REGION is required")
	}

	accessKey := strings.TrimSpace(os.Getenv("AWS_S3_ACCESS_KEY_ID"))
	if accessKey == "" {
		return errors.New("AWS_S3_ACCESS_KEY_ID is required")
	}

	secretKey := strings.TrimSpace(os.Getenv("AWS_S3_SECRET_ACCESS_KEY"))
	if secretKey == "" {
		return errors.New("AWS_S3_SECRET_ACCESS_KEY is required")
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)

	if err != nil {
		return err
	}

	S3Client = s3.NewFromConfig(cfg)

	return nil
}

func UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Read the file content into a buffer
	defer file.Close()

	bucket := strings.TrimSpace(os.Getenv("AWS_BUCKET_NAME"))
	if bucket == "" {
		return "", errors.New("AWS_BUCKET_NAME is required")
	}

	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		return "", errors.New("AWS_REGION is required")
	}

	if S3Client == nil {
		if err := InitS3(); err != nil {
			return "", err
		}
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", err
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := fmt.Sprintf("book-covers/%s-%s",
		uuid.New().String(),
		filepath.Base(fileHeader.Filename),
	)

	_, err = S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &fileName,
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: &contentType,
	})

	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		S3BaseURL,
		bucket,
		region,
		fileName,
	)

	return url, nil
}
