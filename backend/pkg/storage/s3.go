package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Config configures the S3-compatible provider. Endpoint+UsePathStyle let this
// same code target MinIO or any other S3-compatible service, not only AWS.
type S3Config struct {
	Bucket       string
	Region       string
	Endpoint     string // empty = real AWS S3
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
	PublicURL    string // CDN/base URL used to build client-facing links
}

type S3 struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewS3(cfg S3Config) (*S3, error) {
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &S3{client: client, bucket: cfg.Bucket, publicURL: strings.TrimRight(cfg.PublicURL, "/")}, nil
}

func (s *S3) Save(ctx context.Context, folder, filename string, reader io.Reader, size int64, contentType string) (*ObjectMeta, error) {
	ext := filepath.Ext(filename)
	key := filepath.ToSlash(filepath.Join(folder, uuid.NewString()+ext))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 put object: %w", err)
	}

	return &ObjectMeta{Key: key, URL: s.URL(key), Size: size, MimeType: contentType}, nil
}

func (s *S3) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get object: %w", err)
	}
	return out.Body, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3) URL(key string) string {
	if s.publicURL != "" {
		return s.publicURL + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
}
