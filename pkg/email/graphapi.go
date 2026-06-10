package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/clientcredentials"
)

type recipient struct {
	EmailAddress struct {
		Address string `json:"address"`
	} `json:"emailAddress"`
}

type graphMessage struct {
	Message struct {
		Subject string `json:"subject"`
		Body    struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients []recipient `json:"toRecipients"`
	} `json:"message"`
}

type graphAPIClient struct {
	client *http.Client
	userID string
}

func (c *graphAPIClient) Send(report Report, to []string) error {
	msg := graphMessage{}
	msg.Message.Subject = report.Title
	if strings.ToLower(report.Format) == "html" || report.Format == "" {
		msg.Message.Body.ContentType = "HTML"
	} else {
		msg.Message.Body.ContentType = "Text"
	}
	msg.Message.Body.Content = report.Message

	for _, addr := range to {
		r := recipient{}
		r.EmailAddress.Address = addr
		msg.Message.ToRecipients = append(msg.Message.ToRecipients, r)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", c.userID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to send email via Graph API: %s", resp.Status)
	}

	return nil
}

func NewGraphAPIClient(tenant, clientID, clientSecret, userID string) Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant),
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}

	return &graphAPIClient{
		client: config.Client(context.Background()),
		userID: userID,
	}
}
