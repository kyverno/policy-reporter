package email_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/email"
)

type sendMailPayload struct {
	Message struct {
		Subject string `json:"subject"`
		Body    struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients []struct {
			EmailAddress struct {
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"toRecipients"`
		CcRecipients  []json.RawMessage `json:"ccRecipients"`
		BccRecipients []json.RawMessage `json:"bccRecipients"`
	} `json:"message"`
	SaveToSentItems bool `json:"saveToSentItems"`
}

func newGraphServer(t *testing.T, sendMailStatus int, sendMailBody string) (*httptest.Server, *map[string]any) {
	t.Helper()

	captured := map[string]any{}

	mux := http.NewServeMux()
	mux.HandleFunc("/tenant/oauth2/v2.0/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1.0/users/user/sendMail", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured["authorization"] = r.Header.Get("Authorization")
		captured["contentType"] = r.Header.Get("Content-Type")
		captured["body"] = body
		w.WriteHeader(sendMailStatus)
		_, _ = w.Write([]byte(sendMailBody))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, &captured
}

func TestGraphAPIClient_Send(t *testing.T) {
	t.Run("sends expected payload", func(t *testing.T) {
		server, captured := newGraphServer(t, http.StatusAccepted, "")

		client := email.NewGraphAPIClient("tenant", "client", "secret", "user", email.GraphAPIClientOptions{
			CC:                     []string{"cc@example.com"},
			DisableSaveToSentItems: true,
			AzureADEndpoint:        server.URL,
			GraphEndpoint:          server.URL,
		})

		err := client.Send(email.Report{
			Title:   "Report Title",
			Message: "<h1>Summary</h1>",
			Format:  "html",
		}, []string{"to@example.com"})
		assert.NoError(t, err)

		assert.Equal(t, "Bearer test-token", (*captured)["authorization"])
		assert.Equal(t, "application/json", (*captured)["contentType"])

		payload := sendMailPayload{}
		assert.NoError(t, json.Unmarshal((*captured)["body"].([]byte), &payload))
		assert.Equal(t, "Report Title", payload.Message.Subject)
		assert.Equal(t, "HTML", payload.Message.Body.ContentType)
		assert.Equal(t, "<h1>Summary</h1>", payload.Message.Body.Content)
		assert.Len(t, payload.Message.ToRecipients, 1)
		assert.Equal(t, "to@example.com", payload.Message.ToRecipients[0].EmailAddress.Address)
		assert.Len(t, payload.Message.CcRecipients, 1)
		assert.Nil(t, payload.Message.BccRecipients)
		assert.False(t, payload.SaveToSentItems)
	})

	t.Run("uses text content type for non-html reports", func(t *testing.T) {
		server, captured := newGraphServer(t, http.StatusAccepted, "")

		client := email.NewGraphAPIClient("tenant", "client", "secret", "user", email.GraphAPIClientOptions{
			AzureADEndpoint: server.URL,
			GraphEndpoint:   server.URL,
		})

		err := client.Send(email.Report{Title: "Title", Message: "plain", Format: "text"}, []string{"to@example.com"})
		assert.NoError(t, err)

		payload := sendMailPayload{}
		assert.NoError(t, json.Unmarshal((*captured)["body"].([]byte), &payload))
		assert.Equal(t, "Text", payload.Message.Body.ContentType)
		assert.True(t, payload.SaveToSentItems)
	})

	t.Run("returns error including response body on failure", func(t *testing.T) {
		server, _ := newGraphServer(t, http.StatusForbidden, `{"error":{"message":"access denied"}}`)

		client := email.NewGraphAPIClient("tenant", "client", "secret", "user", email.GraphAPIClientOptions{
			AzureADEndpoint: server.URL,
			GraphEndpoint:   server.URL,
		})

		err := client.Send(email.Report{Title: "Title", Message: "msg"}, []string{"to@example.com"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})
}
