package teams_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/target/teams"
)

func TestNewAPI(t *testing.T) {
	h := &http.Client{}

	assert.NotNil(t, teams.NewAPIClient("http://webhook:8080", h))
}
