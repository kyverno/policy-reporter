package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/target/http"
)

var enable = true

type Client interface {
	// Upload given Data the configured AWS storage
	Upload(body *bytes.Buffer, key string) error
}

type s3Client struct {
	bucket               string
	client               *s3.Client
	bucketKeyEnabled     bool
	kmsKeyID             *string
	serverSideEncryption types.ServerSideEncryption
}

type Options func(s *s3Client)

func WithKMS(bucketKeyEnabled bool, kmsKeyID, serverSideEncryption *string) Options {
	return func(s *s3Client) {
		s.bucketKeyEnabled = bucketKeyEnabled
		if *kmsKeyID != "" {
			s.kmsKeyID = kmsKeyID
		}

		if *serverSideEncryption != "" {
			s.serverSideEncryption = types.ServerSideEncryption(s.serverSideEncryption)
		}
	}
}

func (s *s3Client) Upload(body *bytes.Buffer, key string) error {
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(key),
		Body:                 body,
		BucketKeyEnabled:     aws.Bool(s.bucketKeyEnabled),
		SSEKMSKeyId:          s.kmsKeyID,
		ServerSideEncryption: s.serverSideEncryption,
	})
	return err
}

// NewS3Client creates a new S3.client to send Results to S3
func NewS3Client(accessKeyID, secretAccessKey, region, endpoint, bucket string, pathStyle bool, opts ...Options) Client {
	awsConfig, err := createConfig(accessKeyID, secretAccessKey, region)
	if err != nil {
		zap.L().Error("error while creating config", zap.Error(err))
		return nil
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = pathStyle

		if endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
	})

	zap.L().Debug("S3 Client created", zap.String("Region", region), zap.String("Endpoint", endpoint), zap.Bool("PathStyle", pathStyle))

	s3Client := &s3Client{
		bucket: bucket,
		client: client,
	}

	for _, opt := range opts {
		opt(s3Client)
	}

	return s3Client
}

type kinesisClient struct {
	streamName string
	kinesis    *kinesis.Client
}

func (k *kinesisClient) Upload(body *bytes.Buffer, key string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	_, err = k.kinesis.PutRecord(context.TODO(), &kinesis.PutRecordInput{
		StreamName:   aws.String(k.streamName),
		PartitionKey: aws.String(key),
		Data:         data,
	})
	return err
}

// NewKinesisClient creates a new S3.client to send Results to S3
func NewKinesisClient(accessKeyID, secretAccessKey, region, endpoint, streamName string) Client {
	awsConfig, err := createConfig(accessKeyID, secretAccessKey, region)
	if err != nil {
		zap.L().Error("error while creating config", zap.Error(err))
		return nil
	}

	return &kinesisClient{
		streamName,
		kinesis.NewFromConfig(awsConfig, func(o *kinesis.Options) {
			if endpoint != "" {
				o.BaseEndpoint = &endpoint
			}
		}),
	}
}

// NewHubClient creates a new SecurityHub client to send finding events
func NewHubClient(accessKeyID, secretAccessKey, region, endpoint string) *securityhub.Client {
	awsConfig, err := createConfig(accessKeyID, secretAccessKey, region)
	if err != nil {
		zap.L().Error("error while creating config", zap.Error(err))
		return nil
	}

	return securityhub.NewFromConfig(awsConfig, func(o *securityhub.Options) {
		if endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
	})
}

func createConfig(accessKeyID, secretAccessKey, region string) (aws.Config, error) {
	roleARN := os.Getenv("AWS_ROLE_ARN")
	webIdentity := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		if region != "" {
			o.Region = region
		}

		o.HTTPClient = http.NewClient("", false)

		return nil
	})
	if err != nil {
		return aws.Config{}, err
	}

	switch {
	case accessKeyID != "" && secretAccessKey != "":
		zap.L().Debug("configure AWS credentials provider", zap.String("provider", "StaticCredentialsProvider"))
		cfg.Credentials = credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
	case webIdentity != "" && roleARN != "":
		zap.L().Debug("configure AWS credentials provider", zap.String("provider", "WebIdentityRoleProvider"), zap.String("WebIdentityFile", webIdentity))
		cfg.Credentials = stscreds.NewWebIdentityRoleProvider(sts.NewFromConfig(cfg), roleARN, stscreds.IdentityTokenFile(webIdentity))
	default:
		zap.L().Debug("used AWS credentials provider", zap.String("provider", fmt.Sprintf("%T", cfg.Credentials)))
	}

	return cfg, nil
}
