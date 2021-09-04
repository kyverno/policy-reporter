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
	sess AWSSession,
	prefix string,
	bucket string,
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
	
	key := fmt.Sprintf("%s/%s/%s.json", y.prefix, t.Format("2006-01-02"), t.Format(time.RFC3339Nano))
	_, err := s3.New(y.sess).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(y.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := y.client.Do(req)
	helper.HandleHTTPResponse("YANDEXS3", resp, err)
	
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
func NewClient(sess AWSSession, prefix string, bucket string, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {


	
	return &client{
		sess,
		prefix,
		bucket,
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
