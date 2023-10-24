package email

import (
	"fmt"
	"strings"

	mail "github.com/xhit/go-simple-mail/v2"
)

func EncryptionFromString(enc string) mail.Encryption {
	switch strings.ToLower(enc) {
	case "ssl/tls":
		return mail.EncryptionSSLTLS
	case "starttls":
		return mail.EncryptionSTARTTLS
	default:
		return mail.EncryptionNone
	}
}

type Client struct {
	server *mail.SMTPServer
	from   string
}

func (c *Client) Send(report Report, to []string) error {
	if len(to) > 1 {
		c.server.KeepAlive = true
	}

	client, err := c.server.Connect()
	if err != nil {
		return err
	}

	msg := mail.NewMSG().
		SetFrom(fmt.Sprintf("Policy Reporter <%s>", c.from)).
		AddTo(to...).
		SetSubject(report.Title)

	if strings.ToLower(report.Format) == "html" || report.Format == "" {
		msg.SetBody(mail.TextHTML, report.Message)
	} else {
		msg.SetBody(mail.TextPlain, report.Message)
	}

	if msg.Error != nil {
		return msg.Error
	}

	return msg.Send(client)
}

func NewClient(from string, server *mail.SMTPServer) *Client {
	return &Client{server: server, from: from}
}
