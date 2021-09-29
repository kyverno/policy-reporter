package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type client struct {
	sess                  *session.Session
	prefix                string
	bucket                string
	minimumPriority       string
	skipExistingOnStartup bool
}

func (y *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(y.minimumPriority) {
		return
	}

	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(result); err != nil {
		log.Printf("[ERROR] :  Yandex: %v\n", err.Error())
		return
	}
	t := time.Now()
	uploader := s3manager.NewUploader(y.sess)
	key := fmt.Sprintf("%s/%s/%s.json", y.prefix, t.Format("2006-01-02"), t.Format(time.RFC3339Nano))

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(y.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		log.Printf("[ERROR] : Yandex S3 Upload error %v \n", err.Error())
		return
	}

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
func NewClient(AccessKeyID string, SecretAccessKey string, Region string, Endpoint, prefix string, bucket string, minimumPriority string, skipExistingOnStartup bool) target.Client {

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Endpoint:    aws.String(Endpoint),
		Credentials: credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, ""),
	})
	if err != nil {
		log.Printf("[ERROR] : Yandex - %v\n", "Error while creating Yandex Session")
		return nil
	}

	return &client{
		sess,
		prefix,
		bucket,
		minimumPriority,
		skipExistingOnStartup,
	}
}
