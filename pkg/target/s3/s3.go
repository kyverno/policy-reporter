package s3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Options to configure the Kinesis target
type Options struct {
	target.ClientOptions
	S3     helper.AWSClient
	Prefix string
}

type client struct {
	target.BaseClient
	s3     helper.AWSClient
	prefix string
}

func (c *client) Send(result report.Result) {
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(result); err != nil {
		log.Printf("[ERROR] %s : %v\n", c.Name(), err.Error())
		return
	}
	key := fmt.Sprintf("%s/%s/%s-%s-%s.json", c.prefix, result.Timestamp.Format("2006-01-02"), result.Policy, result.ID, result.Timestamp.Format(time.RFC3339Nano))

	err := c.s3.Upload(body, key)
	if err != nil {
		log.Printf("[ERROR] %s : S3 Upload error %v \n", c.Name(), err.Error())
		return
	}

	log.Printf("[INFO] %s PUSH OK", c.Name())
}

// NewClient creates a new S3.client to send Results to S3. It doesnt' work right now
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.S3,
		options.Prefix,
	}
}
