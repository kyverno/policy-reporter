package helper

import (
	"bytes"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/sts"
	"go.uber.org/zap"
)

var enable = true

type AWSClient interface {
	// Upload given Data the configured AWS storage
	Upload(body *bytes.Buffer, key string) error
}

type s3Client struct {
	bucket               string
	uploader             *s3manager.Uploader
	bucketKeyEnabled     *bool
	kmsKeyID             *string
	serverSideEncryption *string
}

type Options func(s *s3Client)

func WithKMS(bucketKeyEnabled *bool, kmsKeyID, serverSideEncryption *string) Options {
	return func(s *s3Client) {
		s.bucketKeyEnabled = bucketKeyEnabled
		if *kmsKeyID != "" {
			s.kmsKeyID = kmsKeyID
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
		SSEKMSKeyId:          s.kmsKeyID,
		ServerSideEncryption: s.serverSideEncryption,
	})
	return err
}

// NewS3Client creates a new S3.client to send Results to S3
func NewS3Client(accessKeyID, secretAccessKey, region, endpoint, bucket string, pathStyle bool, opts ...Options) AWSClient {
	config := createConfig(accessKeyID, secretAccessKey, region, endpoint)
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
	config := createConfig(accessKeyID, secretAccessKey, region, endpoint)

	sess, err := session.NewSession(config)
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
	config := createConfig(accessKeyID, secretAccessKey, region, endpoint)

	sess, err := session.NewSession(config)
	if err != nil {
		zap.L().Error("error while creating SecurityHub session")
		return nil
	}

	optional := make([]*aws.Config, 0)
	if endpoint != "" {
		optional = append(optional, aws.NewConfig().WithEndpoint(endpoint))
	}

	return securityhub.New(sess, optional...)
}

func createConfig(accessKeyID, secretAccessKey, region, endpoint string) *aws.Config {
	baseConfig := &aws.Config{}
	if endpoint != "" {
		baseConfig.Endpoint = aws.String(endpoint)
	}
	if region != "" {
		baseConfig.Region = aws.String(region)
	}

	sess := session.Must(session.NewSession(baseConfig))

	var provider credentials.Provider

	if accessKeyID != "" && secretAccessKey != "" {
		provider = &credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
			},
		}
	} else if os.Getenv("AWS_ROLE_ARN") != "" && os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
		provider = stscreds.NewWebIdentityRoleProvider(
			sts.New(sess),
			os.Getenv("AWS_ROLE_ARN"),
			"",
			os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"),
		)
	} else {
		provider = &ec2rolecreds.EC2RoleProvider{
			Client: ec2metadata.New(sess),
		}
	}

	return &aws.Config{
		Region:                        baseConfig.Region,
		Endpoint:                      baseConfig.Endpoint,
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   credentials.NewCredentials(provider),
	}
}
