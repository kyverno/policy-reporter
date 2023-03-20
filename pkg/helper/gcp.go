package helper

import (
	"bytes"
	"context"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type GCPClient interface {
	// Upload given Data the configured AWS storage
	Upload(body *bytes.Buffer, key string) error
}

type gcsClient struct {
	bucket string
	client *storage.Client
}

func (c *gcsClient) Upload(body *bytes.Buffer, key string) error {
	writer := c.client.Bucket(c.bucket).Object(key).NewWriter(context.Background())
	defer writer.Close()

	_, err := writer.Write(body.Bytes())

	return err
}

// NewS3Client creates a new S3.client to send Results to S3
func NewGCSClient(ctx context.Context, credentials, bucket string) GCPClient {
	cred, err := google.CredentialsFromJSON(ctx, []byte(credentials))
	if err != nil {
		zap.L().Error("error while creating GCS credentials")
		return nil
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(cred))
	if err != nil {
		zap.L().Error("error while creating GCS client")
		return nil
	}

	return &gcsClient{
		bucket,
		client,
	}
}
