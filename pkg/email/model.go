package email

import (
	"context"
)

// Attachment is an in-memory file delivered with an email report.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Report struct {
	Attachments []Attachment
	Title       string
	Message     string
	Format      string
	ClusterName string
}

type Reporter interface {
	Report(ctx context.Context) (Report, error)
}
