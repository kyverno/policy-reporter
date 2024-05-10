package teams

import (
	"net/http"

	teams "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
)

type APIClient interface {
	PostMessage(*adaptivecard.Message) error
}

type apiClient struct {
	webhook string
	client  *teams.TeamsClient
}

func (c *apiClient) PostMessage(message *adaptivecard.Message) error {
	return c.client.Send(c.webhook, message)
}

func NewAPIClient(webhook string, client *http.Client) APIClient {
	msTeams := teams.NewTeamsClient()
	msTeams.SetHTTPClient(client)
	msTeams.SetUserAgent("Policy-Reporter")

	return &apiClient{webhook: webhook, client: msTeams}
}
