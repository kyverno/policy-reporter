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

type client struct {
	target.BaseClient
	s3client helper.S3Client
	prefix   string
}

func (y *client) Send(result *report.Result) {
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(result); err != nil {
		log.Printf("[ERROR] S3 : %v\n", err.Error())
		return
	}
	key := fmt.Sprintf("%s/%s/%s-%s-%s.json", y.prefix, result.Timestamp.Format("2006-01-02"), result.Policy, result.ID, result.Timestamp.Format(time.RFC3339Nano))

	err := y.s3client.Upload(body, key)
	if err != nil {
		log.Printf("[ERROR] S3 : S3 Upload error %v \n", err.Error())
		return
	}

	log.Printf("[INFO] S3 PUSH OK")
}

func (y *client) Name() string {
	return "S3"
}

// NewClient creates a new S3.client to send Results to S3. It doesnt' work right now
func NewClient(s3client helper.S3Client, prefix string, minimumPriority string, sources []string, skipExistingOnStartup bool) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		s3client,
		prefix,
	}
}
