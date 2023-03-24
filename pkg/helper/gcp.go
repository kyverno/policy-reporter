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
	object := c.client.Bucket(c.bucket).Object(key)

	writer := object.NewWriter(context.Background())
	defer writer.Close()

	_, err := writer.Write(body.Bytes())
	if err != nil {
		return err
	}

	return writer.Close()
}

// NewGCSClient creates a new S3.client to send Results to S3
func NewGCSClient(ctx context.Context, credentials, bucket string) GCPClient {
	cred, err := google.CredentialsFromJSON(ctx, []byte(credentials), storage.ScopeReadWrite)
	if err != nil {
		zap.L().Error("error while creating GCS credentials", zap.Error(err))
		return nil
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(cred))
	if err != nil {
		zap.L().Error("error while creating GCS client", zap.Error(err))
		return nil
	}

	return &gcsClient{
		bucket,
		client,
	}
}
