package gcs

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/provider/gcs"
)

// Options to configure the GCS target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Client       gcs.Client
	Prefix       string
}

type client struct {
	target.BaseClient
	customFields map[string]string
	client       gcs.Client
	prefix       string
}

func (c *client) Send(result payload.Payload) {
	if len(c.customFields) > 0 {
		result.AddCustomFields(c.customFields)
	}
	resultBody := result.Body()
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(resultBody); err != nil {
		zap.L().Error(c.Name()+": encode error", zap.Error(err))
		return
	}
	key := result.BlobStorageKey(c.prefix)

	err := c.client.Upload(body, key)
	if err != nil {
		zap.L().Error(c.Name()+": Upload error", zap.Error(err))
		return
	}

	zap.L().Info(c.Name() + ": PUSH OK")
}

func (c *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new GCS.client to send Results to Google Cloud Storage.
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Client,
		options.Prefix,
	}
}
