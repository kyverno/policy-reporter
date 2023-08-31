package email

import (
	"context"
)

type Report struct {
	Title       string
	Message     string
	Format      string
	ClusterName string
}

type Reporter interface {
	Report(ctx context.Context) (Report, error)
}
