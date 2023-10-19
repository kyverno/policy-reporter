package config_test

import (
	"testing"

	"github.com/uptrace/bun/dialect"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/secrets"
)

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
