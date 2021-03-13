package api_test

import (
	"net/http"
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/api"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/discord"
	"github.com/fjogeleit/policy-reporter/pkg/target/loki"
)

func Test_NewServer(t *testing.T) {
	server := api.NewServer(
		report.NewPolicyReportStore(),
		report.NewClusterPolicyReportStore(),
		[]target.Client{
			loki.NewClient("http://localhost:3100", "debug", true, &http.Client{}),
			discord.NewClient("http://webhook:2000", "", false, &http.Client{}),
		},
		8080,
	)

	go server.Start()
}
