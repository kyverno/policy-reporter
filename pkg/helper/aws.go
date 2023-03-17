package helper

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

type AWSClient interface {
	// Upload given Data the configured AWS storage
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

// NewS3Client creates a new S3.client to send Results to S3
func NewS3Client(accessKeyID, secretAccessKey, region, endpoint, bucket string, pathStyle bool) AWSClient {
	config := &aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}
	if pathStyle {
		config.S3ForcePathStyle = &pathStyle
	}

	sess, err := session.NewSession(config)
	if err != nil {
		zap.L().Error("error while creating S3 session")
		return nil
	}

	return &s3Client{
		bucket,
		s3manager.NewUploader(sess),
	}
}

type kinesisClient struct {
	streamName string
	kinesis    *kinesis.Kinesis
}

func (k *kinesisClient) Upload(body *bytes.Buffer, key string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	_, err = k.kinesis.PutRecord(&kinesis.PutRecordInput{
		StreamName:   aws.String(k.streamName),
		PartitionKey: aws.String(key),
		Data:         data,
	})
	return err
}

// NewKinesisClient creates a new S3.client to send Results to S3
func NewKinesisClient(accessKeyID, secretAccessKey, region, endpoint, streamName string) AWSClient {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		zap.L().Error("error while creating Kinesis session")
		return nil
	}

	return &kinesisClient{
		streamName,
		kinesis.New(sess),
	}
}
