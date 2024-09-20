package kinesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
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

func (c *client) Send(result v1alpha2.PolicyReportResult) {
	if len(c.customFields) > 0 {
		props := make(map[string]string, 0)

		for property, value := range c.customFields {
			props[property] = value
		}

		for property, value := range result.Properties {
			props[property] = value
		}

		result.Properties = props
	}

	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(http.NewJSONResult(result)); err != nil {
		zap.L().Error("failed to encode result", zap.String("name", c.Name()), zap.Error(err))
		return
	}
	t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	key := fmt.Sprintf("%s-%s-%s", result.Policy, result.ID, t.Format(time.RFC3339Nano))

	if err := c.kinesis.Upload(body, key); err != nil {
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
