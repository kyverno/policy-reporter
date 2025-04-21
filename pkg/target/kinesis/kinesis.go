package kinesis

import (
	"bytes"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/provider/aws"
)

// Options to configure the Kinesis target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Kinesis      aws.Client
}

type client struct {
	target.BaseClient
	customFields map[string]string
	kinesis      aws.Client
}

func (c *client) Send(result payload.Payload) {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(result.Body()); err != nil {
		zap.L().Error("failed to encode result", zap.String("name", c.Name()), zap.Error(err))
		return
	}

	if err := c.kinesis.Upload(body, result.KinesisKey()); err != nil {
		zap.L().Error("kinesis upload error", zap.String("name", c.Name()), zap.Error(err))
		return
	}

	zap.L().Info("PUSH OK", zap.String("name", c.Name()))
}

func (c *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new Kinesis.client to send Results to AWS Kinesis compatible source
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Kinesis,
	}
}
