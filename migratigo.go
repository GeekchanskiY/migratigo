package migratigo

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	validMigrationNameRegex = `^\d{3}_[a-zA-Z0-9_]+(?:\.up|\.down)\.sql$`
	getMigrationDetailRegex = `^(\d{3})_([a-zA-Z0-9_]+)\.(up|down)\.sql$`
)

type Connector struct {
	migrated         bool
	connection       *sql.DB
	migrationsFS     embed.FS
	migrationsDir    string
	migrationsFilled bool
	Migrations       []Migration
}

type Migration struct {
	Num      int
	Title    string
	Up       bool
	Migrated bool
	Content  string
}

// New creates new migratigo instance
func New(db *sql.DB, migrations embed.FS, migrationsDir string) (*Connector, error) {
	return &Connector{
		migrated:         false,
		connection:       db,
		migrationsFS:     migrations,
		migrationsDir:    migrationsDir,
		migrationsFilled: false,
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

// FillMigrations creates all migrations from embedded sql files
func (c *Connector) FillMigrations() error {
	files, err := fs.ReadDir(c.migrationsFS, c.migrationsDir)
	if err != nil {
		return err
	}

	// name validation and filling migrations
	for _, file := range files {
		if !file.IsDir() {
			err = c.validateMigrationName(file.Name())
			if err != nil {
				return err
			}
			contents, err := fs.ReadFile(c.migrationsFS, filepath.Join(c.migrationsDir, file.Name()))
			if err != nil {
				return err
			}

			num, title, up, err := c.FormatName(file.Name())

			c.Migrations = append(c.Migrations, Migration{
				Num:      num,
				Title:    title,
				Up:       up,
				Migrated: false,
				Content:  string(contents),
			})
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

func (c *Connector) FormatName(filename string) (num int, title string, up bool, err error) {
	regex := regexp.MustCompile(getMigrationDetailRegex)
	matches := regex.FindStringSubmatch(filename)

	// additional check, if validateMigrationName fails
	if len(matches) != 4 {
		return 0, "", false, fmt.Errorf("migration name '%s' is not valid", filename)
	}

	// get all args from migration name
	num, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", false, err
	}

	if num <= 0 || num > 999 {
		return 0, "", false, fmt.Errorf("migration num '%d' is not valid", num)
	}

	title = matches[2]

	if matches[3] == "up" {
		up = true
	} else {
		up = false
	}

	return
}

// Migrate applies migration and creates a db record
func (c *Connector) Migrate() error {
	return nil
}

// Close closes sql connection
func (c *Connector) Close() error {
	return c.connection.Close()
}

func (c *Connector) Connection() (*sql.DB, error) {
	if !c.migrated {
		err := c.FillMigrations()
		if err != nil {
			return nil, err
		}
	}
	return c.connection, nil
}
