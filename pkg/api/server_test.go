package api_test

import (
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
)

func Test_NewServer(t *testing.T) {
	server := api.NewServer(
		report.NewPolicyReportStore(),
		[]target.Client{
			loki.NewClient("http://localhost:3100", "debug", true, &http.Client{}),
			discord.NewClient("http://webhook:2000", "", false, &http.Client{}),
		},
		8080,
	)

	go server.Start()
}
