package target

import (
	"github.com/fjogeleit/policy-reporter/pkg/report"
)

type Client interface {
	Send(result report.Result)
}
