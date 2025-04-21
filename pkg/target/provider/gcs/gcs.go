package gcs

import (
	"bytes"
	"context"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/kyverno/policy-reporter/pkg/http"
)

type Client interface {
	// Upload given Data the configured AWS storage
	Upload(body *bytes.Buffer, key string) error
}

type client struct {
	bucket string
	client *storage.Client
}

func (c *client) Upload(body *bytes.Buffer, key string) error {
	object := c.client.Bucket(c.bucket).Object(key)

	writer := object.NewWriter(context.Background())
	defer writer.Close()

	_, err := writer.Write(body.Bytes())
	if err != nil {
		return err
	}

	return writer.Close()
}

// NewClient creates a new GCS.client to send Results to GCS Bucket
func NewClient(ctx context.Context, credentials, bucket string) Client {
	options := []option.ClientOption{
		option.WithHTTPClient(http.NewClient("", false)),
	}

	if credentials != "" {
		cred, err := google.CredentialsFromJSON(ctx, []byte(credentials), storage.ScopeReadWrite)
		if err != nil {
			zap.L().Error("error while creating GCS credentials", zap.Error(err))
			return nil
		}

		options = append(options, option.WithCredentials(cred))
	}

	baseClient, err := storage.NewClient(ctx, options...)
	if err != nil {
		zap.L().Error("error while creating GCS client", zap.Error(err))
		return nil
	}

	return &client{
		bucket,
		baseClient,
	}
}
