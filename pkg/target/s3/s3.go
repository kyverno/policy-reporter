package s3

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/provider/aws"
)

// Options to configure the S3 target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	S3           aws.Client
	Prefix       string
}

type client struct {
	target.BaseClient
	customFields map[string]string
	s3           aws.Client
	prefix       string
}

func (c *client) Send(result payload.Payload) {
	if len(c.customFields) > 0 {
		if err := result.AddCustomFields(c.customFields); err != nil {
			zap.L().Error(c.Name()+": Error adding custom fields", zap.Error(err))
			return
		}
	}
	resultBody := result.Body()
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(resultBody); err != nil {
		zap.L().Error(c.Name()+": encode error", zap.Error(err))
		return
	}

	key := result.BlobStorageKey(c.prefix)

	if err := c.s3.Upload(body, key); err != nil {
		zap.L().Error(c.Name()+": S3 Upload error", zap.Error(err))
		return
	}

	zap.L().Info(c.Name() + ": PUSH OK")
}

func (c *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new S3.client to send Results to S3.
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.S3,
		options.Prefix,
	}
}
