package s3x

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// awsClient implements Client using the AWS SDK for Go v2.
type awsClient struct {
	s3        *s3.Client
	presigner *s3.PresignClient
}

// New creates a Client backed by the AWS S3 SDK.
// It supports any S3-compatible endpoint (AWS, R2, MinIO, etc.)
// via Config.Endpoint.
func New(cfg Config) (Client, error) {
	endpoint := cfg.Endpoint
	if endpoint != "" {
		scheme := "https"
		if !cfg.UseSSL {
			scheme = "http"
		}
		lower := strings.ToLower(endpoint)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			endpoint = fmt.Sprintf("%s://%s", scheme, endpoint)
		}
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("s3x: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	return &awsClient{
		s3:        client,
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (c *awsClient) PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	_, err := c.s3.PutObject(ctx, input)
	return err
}

func (c *awsClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (c *awsClient) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *awsClient) HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	out, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	info := &ObjectInfo{Key: key}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	if out.LastModified != nil {
		info.LastModified = *out.LastModified
	}
	return info, nil
}

func (c *awsClient) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectEntry, error) {
	out, err := c.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	entries := make([]ObjectEntry, 0, len(out.Contents))
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}
		entries = append(entries, ObjectEntry{Key: *obj.Key, Size: size})
	}
	return entries, nil
}

func (c *awsClient) PresignGetObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	req, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = expiry })
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (c *awsClient) PresignPutObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	req, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) { o.Expires = expiry })
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
