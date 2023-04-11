package helper

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"go.uber.org/zap"
)

type AWSClient interface {
	// Upload given Data the configured AWS storage
	Upload(body *bytes.Buffer, key string) error
}

type s3Client struct {
	bucket               string
	uploader             *s3manager.Uploader
	bucketKeyEnabled     *bool
	kmsKeyId             *string
	serverSideEncryption *string
}

type Options func(s *s3Client)

func WithKMS(bucketKeyEnabled *bool, kmsKeyId, serverSideEncryption *string) Options {
	return func(s *s3Client) {
		s.bucketKeyEnabled = bucketKeyEnabled
		if *kmsKeyId != "" {
			s.kmsKeyId = kmsKeyId
		}

		if *serverSideEncryption != "" {
			s.serverSideEncryption = serverSideEncryption
		}
	}
}

func (s *s3Client) Upload(body *bytes.Buffer, key string) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(key),
		Body:                 body,
		BucketKeyEnabled:     s.bucketKeyEnabled,
		SSEKMSKeyId:          s.kmsKeyId,
		ServerSideEncryption: s.serverSideEncryption,
	})
	return err
}

// NewS3Client creates a new S3.client to send Results to S3
func NewS3Client(accessKeyID, secretAccessKey, region, endpoint, bucket string, pathStyle bool, opts ...Options) AWSClient {
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

	s3Client := &s3Client{
		bucket:   bucket,
		uploader: s3manager.NewUploader(sess),
	}

	for _, opt := range opts {
		opt(s3Client)
	}

	return s3Client
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

// NewHubClient creates a new SecurityHub client to send finding events
func NewHubClient(accessKeyID, secretAccessKey, region, endpoint string) *securityhub.SecurityHub {
	config := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}

	sess, err := session.NewSession(config)
	if err != nil {
		zap.L().Error("error while creating S3 session")
		return nil
	}

	optional := make([]*aws.Config, 0)
	if endpoint != "" {
		optional = append(optional, aws.NewConfig().WithEndpoint(endpoint))
	}

	return securityhub.New(sess, optional...)
}
