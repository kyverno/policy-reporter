package email_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/email"
)

func TestGraphAttachmentsAndErrors(t *testing.T) {
	for _, status := range []int{http.StatusAccepted, http.StatusForbidden, http.StatusRequestEntityTooLarge} {
		for _, enabled := range []bool{false, true} {
			server, captured := newGraphServer(t, status, `{"error":{"message":"local rejection"}}`)
			sender := email.NewGraphAPIClient("tenant", "client", "secret", "user", email.GraphAPIClientOptions{AzureADEndpoint: server.URL, GraphEndpoint: server.URL, CC: []string{"cc@example.test"}, BCC: []string{"bcc@example.test"}, DisableSaveToSentItems: true})
			report := email.Report{Title: "title", Message: "body"}
			data := []byte("Column\n\"雪,quoted\"\n")
			if enabled {
				report.Attachments = []email.Attachment{{Filename: "report.csv", ContentType: "text/csv; charset=utf-8", Data: data}}
			}
			err := sender.Send(report, []string{"to@example.test"})
			if status == http.StatusAccepted {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "local rejection")
			}
			var payload struct {
				Message struct {
					Attachments []struct {
						Type        string `json:"@odata.type"`
						Name        string `json:"name"`
						ContentType string `json:"contentType"`
						Content     []byte `json:"contentBytes"`
					} `json:"attachments"`
					CC  []json.RawMessage `json:"ccRecipients"`
					BCC []json.RawMessage `json:"bccRecipients"`
				} `json:"message"`
				Save bool `json:"saveToSentItems"`
			}
			raw := (*captured)["body"].([]byte)
			require.NoError(t, json.Unmarshal(raw, &payload))
			require.False(t, payload.Save)
			require.Len(t, payload.Message.CC, 1)
			require.Len(t, payload.Message.BCC, 1)
			if enabled {
				require.Len(t, payload.Message.Attachments, 1)
				a := payload.Message.Attachments[0]
				require.Equal(t, data, a.Content)
				require.Equal(t, "#microsoft.graph.fileAttachment", a.Type)
				require.Equal(t, "report.csv", a.Name)
				require.Equal(t, "text/csv; charset=utf-8", a.ContentType)
			} else {
				require.NotContains(t, string(raw), `"attachments"`)
			}
		}
	}
}
