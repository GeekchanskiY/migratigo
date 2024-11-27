package migratigo

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed test_migrations/*
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

	type args struct {
		host     string
		port     string
		username string
		password string
		name     string
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				host:     "localhost",
				port:     "5432",
				username: "postgres",
				password: "postgres",
				name:     "migratigo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection, err := ConnectFromConnectionString(connString)
			if err != nil {
				t.Fatalf("failed to connect: %s", err)
			}
			connector, err := New(connection, testMigrations)
			if (err != nil) != tt.wantErr {
				t.Fatalf("failed to init migratigo: %s", err)
			}

			assert.NotNil(t, connector)
		})
	}
}
