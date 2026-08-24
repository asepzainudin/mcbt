package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/asepzainudin14/mcbt/internal/config"
)

type Client struct {
	s3       *s3.Client
	bucket   string
	maxBytes int64
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})

	c := &Client{
		s3:       client,
		bucket:   cfg.S3Bucket,
		maxBytes: int64(cfg.MaxUploadMB) << 20,
	}

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.bucket)}); err != nil {
		if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.bucket)}); err != nil {
			return nil, fmt.Errorf("create bucket %s: %w", c.bucket, err)
		}
	}

	return c, nil
}

var allowedImageTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/gif":  "gif",
	"image/webp": "webp",
}

func IsAllowedImageType(mime string) bool {
	_, ok := allowedImageTypes[mime]
	return ok
}

func (c *Client) MaxBytes() int64 { return c.maxBytes }

// PutImage stores the reader content under <folder>/<random>.<ext> and returns the object key.
func (c *Client) PutImage(ctx context.Context, r io.Reader, mime, folder string) (string, int64, error) {
	ext, ok := allowedImageTypes[mime]
	if !ok {
		return "", 0, fmt.Errorf("tipe file %s tidak diizinkan", mime)
	}

	key, err := randomKey(folder, ext)
	if err != nil {
		return "", 0, err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return "", 0, fmt.Errorf("membaca file: %w", err)
	}

	_, err = c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mime),
	})
	if err != nil {
		return "", 0, fmt.Errorf("upload ke object storage: %w", err)
	}

	return key, int64(len(data)), nil
}

// GetImage streams the stored object back along with its content type.
func (c *Client) GetImage(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	if err != nil {
		return nil, "", 0, fmt.Errorf("mengambil file dari object storage: %w", err)
	}
	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, contentType, size, nil
}

func randomKey(folder, ext string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.%s", strings.Trim(folder, "/"), hex.EncodeToString(b), ext), nil
}
