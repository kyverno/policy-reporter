package mailgun

import (
	"context"
	"fmt"

	"github.com/mailgun/mailgun-go/v4"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Sender       string
	Mg           mailgun.Mailgun
}

func (c *client) Send(p payload.Payload) {
	emailMsg, err := p.ToEmail()
	if err != nil {
		zap.L().Error(c.Name()+": email conversion error", zap.Error(err))
		return
	}

	for _, recip := range emailMsg.Recipients {
		msg := mailgun.NewMessage(c.sender, emailMsg.Subject, emailMsg.Body, recip)
		ms, id, err := c.mg.Send(context.TODO(), msg)
		if err != nil {
			zap.L().Error(c.Name()+": email sending error", zap.Error(err))
			return
		}
		zap.L().Info(c.Name() + fmt.Sprintf(": email sent to with ID: %s and message: %s\n"+recip, id, ms))
	}
}

type client struct {
	target.BaseClient
	customFields map[string]string
	sender       string
	mg           mailgun.Mailgun
}

func (c *client) Type() target.ClientType {
	return target.SingleSend
}

func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Sender,
		options.Mg,
	}
}
