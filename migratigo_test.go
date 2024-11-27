package migratigo

import (
	"context"
	"embed"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testMigrationsDir    = "test_migrations"
	testBadMigrationsDir = "test_migrations_corrupted"
)

//go:embed test_migrations/*.sql
var testMigrations embed.FS

func TestConnect(t *testing.T) {
	ctx := context.Background()

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")

	if err != nil {
		t.Fatalf("failed to get connection string: %s", connString)
	}

	t.Run("default migrations", func(t *testing.T) {
		connection, err := Connect(connString)
		if err != nil {
			t.Fatalf("failed to connect: %s", err)
		}
		connector, err := New(connection, testMigrations, testMigrationsDir)
		if err != nil {
			t.Fatalf("failed to init migratigo: %s", err)
		}

		assert.NotNil(t, connector)

		err = connector.FillMigrations()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(connector.Migrations))
	})

}

func TestConnector_validateMigrationName(t *testing.T) {
	tests := []struct {
		name          string
		migrationName string
		wantErr       bool
	}{
		{
			name:          "valid migration name",
			migrationName: "001_create_table.up.sql",
			wantErr:       false,
		},
		{
			name:          "valid migration name",
			migrationName: "001_create_table.down.sql",
			wantErr:       false,
		},
		{
			name:          "invalid migration name",
			migrationName: "001_create_tableup.sql",
			wantErr:       true,
		},
		{
			name:          "invalid migration name",
			migrationName: "01_create_table.down.sql",
			wantErr:       true,
		},
		{
			name:          "invalid migration name",
			migrationName: "001.down.sql",
			wantErr:       true,
		},
		{
			name:          "invalid migration name",
			migrationName: "001_.down.sql",
			wantErr:       true,
		},
		{
			name:          "invalid migration name",
			migrationName: "001_create_table.up.html",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connector{}
			err := c.validateMigrationName(tt.migrationName)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
