package gcs

import (
	"bytes"
	"context"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/kyverno/policy-reporter/pkg/target/http"
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
	return newClient(ctx, credentials, bucket)
}

func newClient(ctx context.Context, credentials, bucket string, extraOptions ...option.ClientOption) Client {
	httpClient := http.NewClient("", false)

	var cred *google.Credentials
	if credentials != "" {
		var err error
		cred, err = google.CredentialsFromJSONWithTypeAndParams(ctx, []byte(credentials), google.ServiceAccount, google.CredentialsParams{Scopes: []string{storage.ScopeReadWrite}})
		if err != nil {
			zap.L().Error("error while creating GCS credentials", zap.Error(err))
			return nil
		}
	} else {
		var err error
		cred, err = google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)
		if err != nil {
			zap.L().Warn("no GCS credentials configured and no Application Default Credentials found, uploads are unauthenticated", zap.Error(err))
		}
	}

	// option.WithHTTPClient takes precedence over every other auth option,
	// so the credentials have to be added to the transport of the client itself.
	if cred != nil {
		httpClient.Transport = &oauth2.Transport{
			Source: cred.TokenSource,
			Base:   httpClient.Transport,
		}
	}

	options := append([]option.ClientOption{option.WithHTTPClient(httpClient)}, extraOptions...)

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
