package target

import (
	"context"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"go.uber.org/zap"
)

type TargetType = string

const (
	Loki          TargetType = "Loki"
	Elasticsearch TargetType = "Elasticsearch"
	Slack         TargetType = "Slack"
	Discord       TargetType = "Discord"
	Teams         TargetType = "Teams"
	GoogleChat    TargetType = "GoogleChat"
	Telegram      TargetType = "Telegram"
	Webhook       TargetType = "Webhook"
	S3            TargetType = "S3"
	Kinesis       TargetType = "Kinesis"
	SecurityHub   TargetType = "SecurityHub"
	GCS           TargetType = "GCS"
)

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

func (c *Collection) Update(t *Target) {
	c.mx.Lock()
	c.targets[t.ID] = t
	c.clients = make([]Client, 0)
	c.mx.Unlock()
}

func (c *Collection) Reset(ctx context.Context) {
	clients := c.SyncClients()

	for _, c := range clients {
		if err := c.Reset(ctx); err != nil {
			zap.L().Error("failed to reset target", zap.String("type", c.Type()), zap.String("name", c.Name()))
		}
	}
}

func (c *Collection) Targets() []*Target {
	return helper.ToList(c.targets)
}

func (c *Collection) Clients() []Client {
	if len(c.clients) != 0 {
		return c.clients
	}

	c.clients = helper.MapSlice(c.targets, func(t *Target) Client {
		return t.Client
	})

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
