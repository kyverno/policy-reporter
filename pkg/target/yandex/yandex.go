package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/helper"
)

type client struct {
	s3client              helper.S3Client
	prefix                string
	minimumPriority       string
	skipExistingOnStartup bool
}

func (y *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(y.minimumPriority) {
		return
	}

	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(result); err != nil {
		log.Printf("[ERROR] Yandex : %v\n", err.Error())
		return
	}
	key := fmt.Sprintf("%s/%s/%s.json", y.prefix, result.Timestamp.Format("2006-01-02"), result.Timestamp.Format(time.RFC3339Nano))

	err := y.s3client.Upload(body, key)
	if err != nil {
		log.Printf("[ERROR] Yandex : S3 Upload error %v \n", err.Error())
		return
	}

	log.Printf("[INFO] Yandex PUSH OK")
}

func (y *client) SkipExistingOnStartup() bool {
	return y.skipExistingOnStartup
}

func (y *client) Name() string {
	return "Yandex"
}

func (y *client) MinimumPriority() string {
	return y.minimumPriority
}

// NewClient creates a new Yandex.client to send Results to S3. It doesnt' work right now
func NewClient(s3client helper.S3Client, prefix string, minimumPriority string, skipExistingOnStartup bool) target.Client {
	return &client{
		s3client,
		prefix,
		minimumPriority,
		skipExistingOnStartup,
	}
}
