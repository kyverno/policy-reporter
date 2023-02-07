package kinesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Kinesis target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Kinesis      helper.AWSClient
}

type client struct {
	target.BaseClient
	customFields map[string]string
	kinesis      helper.AWSClient
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
		log.Printf("[ERROR] %s : %v\n", c.Name(), err.Error())
		return
	}
	t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	key := fmt.Sprintf("%s-%s-%s", result.Policy, result.ID, t.Format(time.RFC3339Nano))

	err := c.kinesis.Upload(body, key)
	if err != nil {
		log.Printf("[ERROR] %s : Kinesis Upload error %v \n", c.Name(), err.Error())
		return
	}

	log.Printf("[INFO] %s PUSH OK", c.Name())
}

// NewClient creates a new Kinesis.client to send Results to AWS Kinesis compatible source
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Kinesis,
	}
}
