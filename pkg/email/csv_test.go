package email_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/email"
)

func TestCSVText(t *testing.T) {
	for _, value := range []string{"=1+1", "+cmd", "-1", "@SUM(A1)", "  =1", "\ttext", "\rtext", "\ntext", "\x00=1", "\u00a0+1", "\u200b@x", " \t-1"} {
		t.Run(value, func(t *testing.T) { require.Equal(t, "'"+value, email.CSVText(value)) })
	}
	for _, value := range []string{"", "ordinary", "雪", "a,b", "a\"b", "one\ntwo", " spaces", "already 'quoted"} {
		require.Equal(t, value, email.CSVText(value))
	}
}
