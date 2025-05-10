package target

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type TargetType = string

const (
	Loki          TargetType = "Loki"
	Elasticsearch TargetType = "Elasticsearch"
	Slack         TargetType = "Slack"
	Discord       TargetType = "Discord"
	Teams         TargetType = "Teams"
	GoogleChat    TargetType = "GoogleChat"
	Jira          TargetType = "Jira"
	Telegram      TargetType = "Telegram"
	Webhook       TargetType = "Webhook"
	S3            TargetType = "S3"
	Kinesis       TargetType = "Kinesis"
	SecurityHub   TargetType = "SecurityHub"
	GCS           TargetType = "GCS"
	AlertManager  TargetType = "AlertManager"
)

type Targets struct {
	Loki          *v1alpha1.Config[v1alpha1.LokiOptions]          `mapstructure:"loki"`
	Elasticsearch *v1alpha1.Config[v1alpha1.ElasticsearchOptions] `mapstructure:"elasticsearch"`
	Slack         *v1alpha1.Config[v1alpha1.SlackOptions]         `mapstructure:"slack"`
	Discord       *v1alpha1.Config[v1alpha1.WebhookOptions]       `mapstructure:"discord"`
	Teams         *v1alpha1.Config[v1alpha1.WebhookOptions]       `mapstructure:"teams"`
	Webhook       *v1alpha1.Config[v1alpha1.WebhookOptions]       `mapstructure:"webhook"`
	GoogleChat    *v1alpha1.Config[v1alpha1.WebhookOptions]       `mapstructure:"googleChat"`
	Jira          *v1alpha1.Config[v1alpha1.JiraOptions]          `mapstructure:"jira"`
	Telegram      *v1alpha1.Config[v1alpha1.TelegramOptions]      `mapstructure:"telegram"`
	S3            *v1alpha1.Config[v1alpha1.S3Options]            `mapstructure:"s3"`
	Kinesis       *v1alpha1.Config[v1alpha1.KinesisOptions]       `mapstructure:"kinesis"`
	SecurityHub   *v1alpha1.Config[v1alpha1.SecurityHubOptions]   `mapstructure:"securityHub"`
	GCS           *v1alpha1.Config[v1alpha1.GCSOptions]           `mapstructure:"gcs"`
	AlertManager  *v1alpha1.Config[v1alpha1.AlertManagerOptions]  `mapstructure:"alertManager"`
}

type TargetConfig interface {
	Secret() string
}

type Target struct {
	ID           string
	Type         TargetType
	Client       Client
	ParentConfig TargetConfig
	Config       TargetConfig
}

func (t *Target) Secret() string {
	if t.Config.Secret() != "" {
		return t.Config.Secret()
	}

	return t.ParentConfig.Secret()
}

type Collection struct {
	mx      *sync.Mutex
	clients []Client
	targets map[string]*Target
}

func (c *Collection) AddTarget(key string, t *Target) {
	c.mx.Lock()
	c.targets[key] = t
	c.mx.Unlock()
}

func (c *Collection) RemoveTarget(key string) {
	c.mx.Lock()
	delete(c.targets, key)
	c.mx.Unlock()
}

func (c *Collection) Update(t *Target) {
	c.mx.Lock()
	c.targets[t.ID] = t
	c.clients = make([]Client, 0)
	c.mx.Unlock()
}

func (c *Collection) Reset(ctx context.Context) bool {
	clients := c.SyncClients()

	for _, c := range clients {
		if err := c.Reset(ctx); err != nil {
			zap.L().Error("failed to reset target", zap.String("type", c.Type()), zap.String("name", c.Name()))
		}
	}

	return true
}

func (c *Collection) Targets() []*Target {
	return helper.ToList(c.targets)
}

func (c *Collection) Clients() []Client {
	filterFunc := func(t *Target) Client {
		return t.Client
	}

	c.clients = helper.MapSlice(c.targets, filterFunc)

	return c.clients
}

func (c *Collection) Client(name string) Client {
	return helper.Find(c.Clients(), func(c Client) bool {
		return c.Name() == name
	}, nil)
}

func (c *Collection) SingleSendClients() []Client {
	return helper.Filter(c.Clients(), func(c Client) bool {
		return c.Type() == SingleSend
	})
}

func (c *Collection) SyncClients() []Client {
	return helper.Filter(c.Clients(), func(c Client) bool {
		return c.Type() == SyncSend
	})
}

func (c *Collection) BatchSendClients() []Client {
	return helper.Filter(c.Clients(), func(c Client) bool {
		return c.Type() == BatchSend
	})
}

func (c *Collection) UsesSecrets() bool {
	useSecrets := helper.Filter(c.Targets(), func(t *Target) bool {
		return t.Secret() != ""
	})

	return len(useSecrets) > 0
}

func (c *Collection) Empty() bool {
	return c.Length() == 0
}

func (c *Collection) Length() int {
	return len(c.targets)
}

func NewCollection(targets ...*Target) *Collection {
	collection := &Collection{
		clients: make([]Client, 0),
		targets: make(map[string]*Target, 0),
		mx:      new(sync.Mutex),
	}

	for _, t := range targets {
		if t != nil {
			collection.Update(t)
		}
	}

	return collection
}
