package migratigo

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"regexp"

	_ "github.com/lib/pq"
)

const (
	validMigrationNameRegex = `^\d{3}_[a-zA-Z0-9_]+(?:\.up|\.down)\.sql$`
)

type Connector struct {
	migrated     bool
	connection   *sql.DB
	migrationsFS embed.FS
}

func New(db *sql.DB, migrations embed.FS) (*Connector, error) {
	return &Connector{
		migrated:     false,
		connection:   db,
		migrationsFS: migrations,
	}, nil
}

// Connect connects to sql db from connection string
func Connect(connString string) (*sql.DB, error) {
	connection, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = connection.Ping()
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func (c *Connector) RunMigrations() error {
	files, err := fs.ReadDir(c.migrationsFS, "test_migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			err = c.validateMigrationName(file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateMigrationName checks if file names are in format xxx_name.up/down.sql
func (c *Connector) validateMigrationName(name string) error {
	regex := regexp.MustCompile(validMigrationNameRegex)

	if !regex.MatchString(name) {
		return fmt.Errorf("migration name '%s' is not valid", name)
	}

	return nil
}

func (c *Connector) Close() error {
	return c.connection.Close()
}

func (c *Connector) Connection() (*sql.DB, error) {
	if !c.migrated {
		err := c.RunMigrations()
		if err != nil {
			return nil, err
		}
	}
	return c.connection, nil
}
