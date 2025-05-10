package alertmanager

import (
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Client extends the target.Client interface with AlertManager-specific functionality
type Client interface {
	target.Client
}

// Ensure the client type implements the Client interface
var _ Client = (*client)(nil)
