package payload

import (
	"fmt"
	"strings"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/slack-go/slack"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload/scutils"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
)

type EmailMsg struct {
	Recipients []string `json:"recipients"`
	Attachment []byte   `json:"attachment"`
	CC         []string `json:"cc"`
	Bcc        []string `json:"bcc"`
	Body       string   `json:"body"`
	Subject    string   `json:"subject"`
}

type Payload interface {
	// Return a unique identifier to the payload
	GetID() string
	// Get the JSON representation of the payload
	Body() interface{}
	// Get the payload as a loki stream
	ToLoki() (Stream, error)
	// Get the Telegram notification string
	ToTelegram(chatId string) (string, error)
	// Get the Teams representation
	ToTeams() (adaptivecard.Container, error)
	// Get the Slack representation
	ToSlack(channel string) *slack.Attachment
	// Get the Discord representation
	ToDiscord() DiscordPayload
	// Get the google chat representation
	ToGoogleChat() (*GCPayload, error)
	// Get the Email representation
	ToEmail() (EmailMsg, error)
	// Get the AWS security finding
	ToSecurityHubFindings(scutils.SecurityHubConfig) (*types.AwsSecurityFinding, error)
	// Get the key in a blob storage that this payload should be pushed to
	BlobStorageKey(string) string
	// Get the Kinesis key the payload should be pushed to
	KinesisKey() string
	// Add any custom key value pairs to the payload
	AddCustomFields(map[string]string) error
}

type PolicyReportResultPayload struct {
	Result v1alpha2.PolicyReportResult
}

func (p *PolicyReportResultPayload) GetID() string {
	return p.Result.GetID()
}

func (p *PolicyReportResultPayload) AddCustomFields(fieldMap map[string]string) error {
	props := make(map[string]string, 0)

	for property, value := range fieldMap {
		props[property] = value
	}

	for property, value := range p.Result.Properties {
		props[property] = value
	}

	p.Result.Properties = props
	return nil
}

func (p *PolicyReportResultPayload) BlobStorageKey(prefix string) string {
	t := time.Unix(p.Result.Timestamp.Seconds, int64(p.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s/%s/%s-%s-%s.json", prefix, t.Format("2006-01-02"), p.Result.Policy, p.Result.ID, t.Format(time.RFC3339Nano))
}

func (p *PolicyReportResultPayload) Body() interface{} {
	return http.NewJSONResult(p.Result)
}

func (s *PolicyReportResultPayload) KinesisKey() string {
	t := time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s-%s-%s", s.Result.Policy, s.Result.ID, t.Format(time.RFC3339Nano))
}
