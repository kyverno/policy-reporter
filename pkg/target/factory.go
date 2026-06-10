package target

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
)

type Factory interface {
	CreateClients(config *Targets) *Collection
	CreateLokiTarget(config, parent *targetconfig.Config[v1alpha1.LokiOptions]) *Target
	CreateSingleClient(*v1alpha1.TargetConfig) (*Target, error)
	CreateElasticsearchTarget(config, parent *targetconfig.Config[v1alpha1.ElasticsearchOptions]) *Target
	CreateSlackTarget(config, parent *targetconfig.Config[v1alpha1.SlackOptions]) *Target
	CreateDiscordTarget(config, parent *targetconfig.Config[v1alpha1.WebhookOptions]) *Target
	CreateTeamsTarget(config, parent *targetconfig.Config[v1alpha1.WebhookOptions]) *Target
	CreateWebhookTarget(config, parent *targetconfig.Config[v1alpha1.WebhookOptions]) *Target
	CreateTelegramTarget(config, parent *targetconfig.Config[v1alpha1.TelegramOptions]) *Target
	CreateGoogleChatTarget(config, parent *targetconfig.Config[v1alpha1.WebhookOptions]) *Target
	CreateJiraTarget(config, parent *targetconfig.Config[v1alpha1.JiraOptions]) *Target
	CreateS3Target(config, parent *targetconfig.Config[v1alpha1.S3Options]) *Target
	CreateKinesisTarget(config, parent *targetconfig.Config[v1alpha1.KinesisOptions]) *Target
	CreateSecurityHubTarget(config, parent *targetconfig.Config[v1alpha1.SecurityHubOptions]) *Target
	CreateGCSTarget(config, parent *targetconfig.Config[v1alpha1.GCSOptions]) *Target
	CreateSplunkTarget(config, parent *targetconfig.Config[v1alpha1.SplunkOptions]) *Target
}
