package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2/clientcredentials"
)

type emailAddress struct {
	Address string `json:"address"`
	// Name is the optional display name shown in email clients.
	Name string `json:"name,omitempty"`
}

type recipient struct {
	EmailAddress emailAddress `json:"emailAddress"`
}

type graphMessage struct {
	Message struct {
		Subject string `json:"subject"`
		Body    struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients  []recipient `json:"toRecipients"`
		CcRecipients  []recipient `json:"ccRecipients,omitempty"`
		BccRecipients []recipient `json:"bccRecipients,omitempty"`
	} `json:"message"`
	// SaveToSentItems controls whether the sent message is saved in the Sent Items folder.
	// Always explicitly sent to avoid relying on server-side defaults.
	SaveToSentItems bool `json:"saveToSentItems"`
}

type graphAPIClient struct {
	oauthConfig            *clientcredentials.Config
	userID                 string
	cc                     []string
	bcc                    []string
	disableSaveToSentItems bool
	graphEndpoint          string
	// ctx is used for OAuth2 token fetching and HTTP requests.
	// TODO: propagate a caller-supplied context once the Sender interface accepts one.
	ctx context.Context
}

// makeRecipients converts a list of email address strings into Graph API recipient objects.
// Returns nil (not an empty slice) when addrs is empty so that fields tagged with
// omitempty are correctly omitted from the JSON payload.
func makeRecipients(addrs []string) []recipient {
	if len(addrs) == 0 {
		return nil
	}
	rs := make([]recipient, 0, len(addrs))
	for _, addr := range addrs {
		r := recipient{}
		r.EmailAddress.Address = addr
		rs = append(rs, r)
	}
	return rs
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
	msg.Message.ToRecipients = makeRecipients(to)
	msg.Message.CcRecipients = makeRecipients(c.cc)
	msg.Message.BccRecipients = makeRecipients(c.bcc)
	msg.SaveToSentItems = !c.disableSaveToSentItems

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Create an HTTP client that obtains OAuth2 tokens lazily per request.
	httpClient := c.oauthConfig.Client(c.ctx)

	url := fmt.Sprintf("%s/v1.0/users/%s/sendMail", c.graphEndpoint, c.userID)
	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send email via Graph API: %s – %s", resp.Status, string(respBody))
	}

	return nil
}

// GraphAPIClientOptions holds optional settings for the Microsoft Graph API email sender.
type GraphAPIClientOptions struct {
	// CC is a list of email addresses to carbon-copy on every sent message.
	CC []string
	// BCC is a list of email addresses to blind-carbon-copy on every sent message.
	BCC []string
	// DisableSaveToSentItems controls whether sent messages are saved in the Sent Items folder.
	// Defaults to false (which means messages ARE saved). Set to true to suppress saving.
	DisableSaveToSentItems bool
	// AzureADEndpoint overrides the default Azure Active Directory / Entra ID endpoint.
	// Defaults to "https://login.microsoftonline.com" if not set.
	AzureADEndpoint string
	// GraphEndpoint overrides the default Microsoft Graph API endpoint.
	// Defaults to "https://graph.microsoft.com" if not set.
	GraphEndpoint string
}

// NewGraphAPIClient creates a Sender backed by the Microsoft Graph API.
// OAuth2 tokens are obtained lazily on each Send call using the client credentials flow.
// Pass GraphAPIClientOptions to configure CC, BCC, and Sent Items behaviour.
func NewGraphAPIClient(tenant, clientID, clientSecret, userID string, opts GraphAPIClientOptions) Sender {
	azureADEndpoint := strings.TrimSuffix(opts.AzureADEndpoint, "/")
	if azureADEndpoint == "" {
		azureADEndpoint = "https://login.microsoftonline.com"
	}
	graphEndpoint := strings.TrimSuffix(opts.GraphEndpoint, "/")
	if graphEndpoint == "" {
		graphEndpoint = "https://graph.microsoft.com"
	}

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s/%s/oauth2/v2.0/token", azureADEndpoint, tenant),
		Scopes:       []string{fmt.Sprintf("%s/.default", graphEndpoint)},
	}

	return &graphAPIClient{
		oauthConfig:            config,
		userID:                 userID,
		cc:                     opts.CC,
		bcc:                    opts.BCC,
		disableSaveToSentItems: opts.DisableSaveToSentItems,
		graphEndpoint:          graphEndpoint,
		ctx:                    context.Background(),
	}
}
