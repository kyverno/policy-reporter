package s3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Kinesis target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	S3           helper.AWSClient
	Prefix       string
}

type client struct {
	target.BaseClient
	customFields map[string]string
	s3           helper.AWSClient
	prefix       string
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
		c.Logger().Error(c.Name()+": encode error", zap.Error(err))
		return
	}
	t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	key := fmt.Sprintf("%s/%s/%s-%s-%s.json", c.prefix, t.Format("2006-01-02"), result.Policy, result.ID, t.Format(time.RFC3339Nano))

	err := c.s3.Upload(body, key)
	if err != nil {
		c.Logger().Error(c.Name()+": S3 Upload error", zap.Error(err))
		return
	}

	c.Logger().Info(c.Name() + ": PUSH OK")
}

// NewClient creates a new S3.client to send Results to S3. It doesnt' work right now
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.S3,
		options.Prefix,
	}
}
