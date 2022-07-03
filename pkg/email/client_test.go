package email_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/email"
	mail "github.com/xhit/go-simple-mail/v2"
)

func Test_EncryptionFromString(t *testing.T) {
	t.Run("EncryptionFromString.SSLTLS", func(t *testing.T) {
		encryption := email.EncryptionFromString("ssl/tls")
		if encryption != mail.EncryptionSSLTLS {
			t.Errorf("Unexpected encryption mapping: %d", encryption)
		}
	})
	t.Run("EncryptionFromString.STARTTLS", func(t *testing.T) {
		encryption := email.EncryptionFromString("starttls")
		if encryption != mail.EncryptionSTARTTLS {
			t.Errorf("Unexpected encryption mapping: %d", encryption)
		}
	})
	t.Run("EncryptionFromString.Default", func(t *testing.T) {
		encryption := email.EncryptionFromString("")
		if encryption != mail.EncryptionNone {
			t.Errorf("Unexpected encryption mapping: %d", encryption)
		}
	})
}

func Test_NewClient(t *testing.T) {
	client := email.NewClient("policy-reporter@kyverno.io", nil)
	if client == nil {
		t.Errorf("Unexpected client result")
	}
}
