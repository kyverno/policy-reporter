package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/uptrace/bun/dialect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
)

const (
	secretName    = "secret-values"
	mountedSecret = "/tmp/secrets-9999"
)

func newFakeClient() v1.SecretInterface {
	return k8sfake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"host":     []byte("http://localhost:9200"),
			"username": []byte("username"),
			"password": []byte("password"),
			"apiKey":   []byte("apiKey"),
			"database": []byte("database"),
			"dsn":      []byte(""),
		},
	}).CoreV1().Secrets("default")
}

func mountSecret() {
	secretValues := secrets.Values{
		Host:     "http://localhost:9200",
		Username: "username",
		Password: "password",
		Database: "database",
		DSN:      "",
	}
	file, _ := json.MarshalIndent(secretValues, "", " ")
	_ = os.WriteFile(mountedSecret, file, 0o644)
}

func Test_ResolveDatabase(t *testing.T) {
	factory := config.NewDatabaseFactory(nil)

	t.Run("SQLite Fallback", func(t *testing.T) {
		db := factory.NewSQLite("test.db")
		if db == nil || db.Dialect().Name() != dialect.SQLite {
			t.Error("Expected SQLite database as fallback")
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		db := factory.NewMySQL(config.Database{
			Username:  "admin",
			Password:  "password",
			Host:      "localhost:3306",
			EnableSSL: true,
		})
		if db == nil || db.Dialect().Name() != dialect.MySQL {
			t.Error("Expected MySQL DB")
		}
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		db := factory.NewPostgres(config.Database{
			Username:  "admin",
			Password:  "password",
			Host:      "localhost:5432",
			EnableSSL: true,
		})
		if db == nil || db.Dialect().Name() != dialect.PG {
			t.Error("Expected PostgreSQL")
		}
	})
}

func Test_DatabaseValuesFromSecret(t *testing.T) {
	factory := config.NewDatabaseFactory(secrets.NewClient(newFakeClient()))
	mountSecret()

	t.Run("Values from SecretRef", func(t *testing.T) {
		db := factory.NewPostgres(config.Database{SecretRef: secretName, EnableSSL: false})
		if db == nil {
			t.Error("Expected PostgreSQL connection created")
		}
	})

	t.Run("Values from MountedSecret", func(t *testing.T) {
		db := factory.NewMySQL(config.Database{MountedSecret: mountedSecret, EnableSSL: false})
		if db == nil {
			t.Error("Expected MySQL connection created")
		}
	})

	t.Run("Get none existing mounted secret skips target", func(t *testing.T) {
		db := factory.NewPostgres(config.Database{MountedSecret: "no-exists"})
		if db != nil {
			t.Error("Expected no connection created without host or DSN config")
		}
	})
}
