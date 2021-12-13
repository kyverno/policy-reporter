package helper

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client interface {
	Upload(body *bytes.Buffer, key string) error
}

type s3Client struct {
	bucket   string
	uploader *s3manager.Uploader
}

func (s *s3Client) Upload(body *bytes.Buffer, key string) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

// NewClient creates a new S3.client to send Results to S3. It doesnt' work right now
func NewClient(accessKeyID, secretAccessKey, region, endpoint, bucket string) S3Client {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Printf("[ERROR]: %v\n", "Error while creating S3 Session")
		return nil
	}

	return &s3Client{
		bucket,
		s3manager.NewUploader(sess),
	}
}
