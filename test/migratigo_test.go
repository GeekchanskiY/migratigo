package test

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/GeekchanskiY/migratigo/pkg/migratigo"
	"github.com/GeekchanskiY/migratigo/pkg/migration"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testMigrationsDir = "test_migrations"
)

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

	connection, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("failed to open connection: %s", err)
	}
	defer connection.Close()

	t.Run("default migrations", func(t *testing.T) {
		connector, err := migratigo.New(connection, testMigrationsDir)
		if err != nil {
			t.Fatalf("failed to init migratigo: %s", err)
		}

		assert.NotNil(t, connector)

		err = connector.RunMigrations(false)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(connector.Migrations))

		rows, err := connection.Query(`select * from migrations;`)
		assert.NoError(t, err)

		migrations := make([]migration.Migration, 0, 4)
		for rows.Next() {
			var m migration.Migration
			err = rows.Scan(&m.Num, &m.Title, &m.Migrated)
			assert.NoError(t, err)

			migrations = append(migrations, m)
		}
		assert.Equal(t, len(connector.Migrations)/2, len(migrations))
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
			_, err := migration.FromFile(filepath.Join(testMigrationsDir, tt.migrationName))
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
