package aws_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target/provider/aws"
)

func TestS3Client(t *testing.T) {
	client := aws.NewS3Client("access", "secret", "eu-central-1", "http://s3.aws.com", "policy-reporter", false, aws.WithKMS(true, helper.ToPointer("kms"), helper.ToPointer("encryption")))

	assert.NotNil(t, client)
}

func TestKinesisClient(t *testing.T) {
	client := aws.NewKinesisClient("access", "secret", "eu-central-1", "http://kinesis.aws.com", "policy-reporter")

	assert.NotNil(t, client)
}

func TestSecurityHubClient(t *testing.T) {
	client := aws.NewHubClient("access", "secret", "eu-central-1", "http://securityhub.aws.com")

	assert.NotNil(t, client)
}
