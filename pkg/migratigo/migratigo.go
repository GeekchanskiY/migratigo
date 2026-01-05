// Copyright 2025 GeekchanskiY
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migratigo

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/GeekchanskiY/migratigo/pkg/migration"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Connector struct {
	migrated         bool
	connection       *sql.DB
	migrationsDir    string
	migrationsFilled bool
	Migrations       []*migration.Migration
}

//go:embed migratigo.sql
var schemaMigrations string

// New creates new migratigo instance, does initial duty
func New(db *sql.DB, migrationsDir string) (*Connector, error) {
	return &Connector{
		migrated:         false,
		connection:       db,
		migrationsDir:    migrationsDir,
		migrationsFilled: false,
	}, nil
}

func NewFromSqlx(db sqlx.DB, migrationsDir string) (*Connector, error) {
	return &Connector{
		migrated:         false,
		connection:       db.DB,
		migrationsDir:    migrationsDir,
		migrationsFilled: false,
	}, nil
}

func InitTable(conn *sql.DB) error {
	_, err := conn.Exec(schemaMigrations)
	if err != nil {
		return err
	}

	return nil
}

// FillMigrations creates all migrations from embedded sql files, and validates them
func (c *Connector) fillMigrations(noOpposite bool) error {
	files, err := os.ReadDir(c.migrationsDir)
	if err != nil {
		return err
	}

	// name validation and filling migrations
	for _, file := range files {
		if !file.IsDir() {
			newMigration, err := migration.FromFile(filepath.Join(c.migrationsDir, file.Name()))
			if err != nil {
				return err
			}

			c.Migrations = append(c.Migrations, newMigration)
		}
	}

	sort.Slice(c.Migrations, func(i, j int) bool {
		if c.Migrations[i].Num == c.Migrations[j].Num {
			return c.Migrations[i].Up == true
		}
		return c.Migrations[i].Num < c.Migrations[j].Num
	})

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

	for i, m := range c.Migrations {
		err := c.migrate(m)
		if err != nil {
			return err
		}
		c.Migrations[i].Migrated = true
	}

	return nil
}

// migrate applies Migration and creates a db record
func (c *Connector) migrate(migration *migration.Migration) error {
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

func (c *Connector) checkIfMigrationExists(m *migration.Migration) (bool, error) {
	q := `SELECT exists(SELECT * FROM migrations WHERE num = $1) `

	var count bool

	err := c.connection.QueryRow(q, m.Num).Scan(&count)
	if err != nil {
		return false, err
	}

	return count, nil
}

func (c *Connector) applyMigration(m *migration.Migration) error {
	_, err := c.connection.Exec(m.Content)
	return err
}

func (c *Connector) confirmMigration(m *migration.Migration) error {
	q := `INSERT INTO migrations(num, title, applied) VALUES ($1, $2, $3);`

	_, err := c.connection.Exec(q, m.Num, m.Title, m.Up)
	if err != nil {
		return err
	}

	return nil
}
