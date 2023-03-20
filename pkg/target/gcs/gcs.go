package gcs

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

// Options to configure the GCS target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Client       helper.GCPClient
	Prefix       string
}

type client struct {
	target.BaseClient
	customFields map[string]string
	client       helper.GCPClient
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
		zap.L().Error(c.Name()+": encode error", zap.Error(err))
		return
	}
	t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	key := fmt.Sprintf("%s/%s/%s-%s-%s.json", c.prefix, t.Format("2006-01-02"), result.Policy, result.ID, t.Format(time.RFC3339Nano))

	err := c.client.Upload(body, key)
	if err != nil {
		zap.L().Error(c.Name()+": Upload error", zap.Error(err))
		return
	}

	zap.L().Info(c.Name() + ": PUSH OK")
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
