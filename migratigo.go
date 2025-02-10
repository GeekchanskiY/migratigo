package migratigo

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	validMigrationNameRegex = `^\d{3}_[a-zA-Z0-9_]+(?:\.up|\.down)\.sql$`
	getMigrationDetailRegex = `^(\d{3})_([a-zA-Z0-9_]+)\.(up|down)\.sql$`
)

type Connector struct {
	log              *zap.Logger
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

//go:embed migratigo.sql
var schemaMigrations embed.FS

// New creates new migratigo instance, does initial duty
func New(db *sql.DB, migrations embed.FS, migrationsDir string, logger *zap.Logger) (*Connector, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	connector := Connector{
		migrated:         false,
		connection:       db,
		migrationsFS:     migrations,
		migrationsDir:    migrationsDir,
		migrationsFilled: false,
		log:              logger,
	}

	return &connector, nil
}

func NewFromSqlx(db sqlx.DB, migrations embed.FS, migrationsDir string, logger *zap.Logger) (*Connector, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	connector := Connector{
		migrated:         false,
		connection:       db.DB,
		migrationsFS:     migrations,
		migrationsDir:    migrationsDir,
		migrationsFilled: false,
		log:              logger,
	}

	return &connector, nil
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

// FillMigrations creates all migrations from embedded sql files, and validates them
func (c *Connector) fillMigrations(noOpposite bool) error {
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

			num, title, up, err := c.formatName(file.Name())

			c.Migrations = append(c.Migrations, Migration{
				Num:      num,
				Title:    title,
				Up:       up,
				Migrated: false,
				Content:  string(contents),
			})
		}
	}

	sort.Slice(c.Migrations, func(i, j int) bool {
		if c.Migrations[i].Num == c.Migrations[j].Num {
			return c.Migrations[i].Up == true
		}
		return c.Migrations[i].Num < c.Migrations[j].Num
	})

	c.log.Debug("migrations filled successfully", zap.Int("migrations", len(c.Migrations)))
	c.log.Debug("validating migrations")

	found := false
	for origNum, migrationOrig := range c.Migrations {

		for foundNum, migrationFound := range c.Migrations {

			if migrationFound.Num == migrationOrig.Num && origNum != foundNum {

				if migrationFound.Up == migrationOrig.Up {
					upStr := "down"

					if migrationFound.Up {
						upStr = "up"
					}

					return fmt.Errorf("migration %d has 2 %s files", migrationOrig.Num, upStr)
				}

				if found {
					return fmt.Errorf("found 2 same migrations: %d", migrationOrig.Num)
				}

				found = true
			}
		}

		if !found && !noOpposite {
			return fmt.Errorf("migration %d not found in any opposite migrations", migrationOrig.Num)
		}

		found = false
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

func (c *Connector) formatName(filename string) (num int, title string, up bool, err error) {
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

// RunMigrations fills and runs migrations. if noOpposite flag provided, you can not create .down migrations.
func (c *Connector) RunMigrations(noOpposite bool) error {
	err := c.fillMigrations(noOpposite)
	if err != nil {
		return err
	}

	return c.runMigrations()
}

// runMigrations iterates through all migrations and runs them
func (c *Connector) runMigrations() error {
	if len(c.Migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}

	schemaMigrationsContent, err := fs.ReadFile(schemaMigrations, "migratigo.sql")
	if err != nil {
		return err
	}

	_, err = c.connection.Exec(string(schemaMigrationsContent))
	if err != nil {
		return err
	}

	for i, migration := range c.Migrations {
		err := c.migrate(migration)
		if err != nil {
			return err
		}
		c.Migrations[i].Migrated = true
	}

	return nil
}

// migrate applies migration and creates a db record
func (c *Connector) migrate(migration Migration) error {
	exists, err := c.checkIfMigrationExists(migration)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	err = c.applyMigration(migration)
	if err != nil {
		return err
	}

	err = c.confirmMigration(migration)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) checkIfMigrationExists(migration Migration) (bool, error) {
	q := `SELECT exists(SELECT * FROM migrations WHERE num = $1) `

	var count bool

	err := c.connection.QueryRow(q, migration.Num).Scan(&count)
	if err != nil {
		return false, err
	}

	return count, nil
}

func (c *Connector) applyMigration(migration Migration) error {
	_, err := c.connection.Exec(migration.Content)
	return err
}

func (c *Connector) confirmMigration(migration Migration) error {
	q := `INSERT INTO migrations(num, title, applied) VALUES ($1, $2, $3);`

	_, err := c.connection.Exec(q, migration.Num, migration.Title, migration.Up)
	if err != nil {
		return err
	}
	return nil
}

// Close closes sql connection
func (c *Connector) Close() error {
	return c.connection.Close()
}
