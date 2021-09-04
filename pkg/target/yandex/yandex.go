package yandex

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"


	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/helper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}



type client struct {
	
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func (y *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(s.minimumPriority) {
		return
	}

	payload := newPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] YandexS3 : %v\n", err.Error())
		return
	}
	/*

	TDB
	*/

}

func (s *client) SkipExistingOnStartup() bool {
	return s.skipExistingOnStartup
}

func (s *client) Name() string {
	return "YandexS3"
}

func (s *client) MinimumPriority() string {
	return s.minimumPriority
}

/

// NewClient creates a new Yandex.client to send Results to S3. It doesnt' work right now
func NewClient(sess, minimumPriority string, skipExistingOnStartup bool, s3client s3client) target.Client {


	
	return &client{
		sess,
		minimumPriority,
		skipExistingOnStartup,
		s3client,
	}
}
