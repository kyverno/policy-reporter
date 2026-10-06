package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	db "github.com/kyverno/policy-reporter/pkg/database"
)

func New(store *db.Store) *server.StreamableHTTPServer {
	s := server.NewMCPServer(
		"Policy Reporter",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerTools(s, store)

	return server.NewStreamableHTTPServer(s, server.WithEndpointPath("mcp"))
}
