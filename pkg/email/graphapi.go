package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// requestTimeout bounds each HTTP request (token fetch and sendMail),
// including connection setup, TLS handshake and reading the response body.
const requestTimeout = 30 * time.Second

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
	// httpClient handles OAuth2 tokens transparently, fetching and refreshing
	// them as needed via the client credentials flow.
	httpClient             *http.Client
	userID                 string
	cc                     []string
	bcc                    []string
	disableSaveToSentItems bool
	graphEndpoint          string
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

	url := fmt.Sprintf("%s/v1.0/users/%s/sendMail", c.graphEndpoint, c.userID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
// OAuth2 tokens are obtained via the client credentials flow on first use and
// refreshed automatically when they expire.
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

	// Token requests are made through the client stored in the oauth2.HTTPClient
	// context value, so give it a timeout too instead of http.DefaultClient's none.
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Timeout: requestTimeout})
	httpClient := config.Client(ctx)
	httpClient.Timeout = requestTimeout

	return &graphAPIClient{
		httpClient:             httpClient,
		userID:                 userID,
		cc:                     opts.CC,
		bcc:                    opts.BCC,
		disableSaveToSentItems: opts.DisableSaveToSentItems,
		graphEndpoint:          graphEndpoint,
	}
}
