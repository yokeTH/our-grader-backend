package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Config struct {
	BucketName      string `env:"BUCKET_NAME"`
	AccountID       string `env:"ACCOUNT_ID"`
	AccessKeyID     string `env:"ACCESS_KEY_ID"`
	AccessKeySecret string `env:"ACCESS_KEY_SECRET"`
	UrlFormat       string `env:"URL_FORMAT"`
}

type R2Storage struct {
	*Storage
	config R2Config
}

func NewR2Storage(r2Config R2Config) (IStorage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKeyID, r2Config.AccessKeySecret, "")),
		config.WithRegion("auto"),
		config.WithBaseEndpoint(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2Config.AccountID)),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)
	return &R2Storage{&Storage{presignClient, client, r2Config.BucketName}, r2Config}, nil
}

func (s *R2Storage) UploadFile(ctx context.Context, key string, contentType string, file io.Reader) error {
	// Calculate the content length (if it's available in the stream).
	// If the file's size is known beforehand, this can be useful.
	// You can also try to set an explicit "Content-Length" header if needed.
	// Here we assume you have the file size available.
	fileSize, err := getFileSize(file) // This function should calculate the size if needed.
	if err != nil {
		return fmt.Errorf("unable to determine file size: %w", err)
	}

	// Use the file data and upload it to R2 without specifying CRC32 checksum.
	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.BucketName),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		Body:          file,
		ContentLength: aws.Int64(fileSize), // Optional if size is known
	})

	return err
}

func (s *R2Storage) GetSignedUrl(ctx context.Context, key string, expires time.Duration) (string, error) {
	req, err := s.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (s *R2Storage) GetPublicUrl(key string) string {
	return fmt.Sprintf(s.config.UrlFormat, key)
}

func (s *R2Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})

	return err
}

func (s *R2Storage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})

	return obj.Body, err
}

func getFileSize(file io.Reader) (int64, error) {
	// Read the entire file into a buffer to get the size.
	buf := make([]byte, 0)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return 0, err
	}
	return int64(n), nil
}
