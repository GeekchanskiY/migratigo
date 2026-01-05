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

package config

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/GeekchanskiY/migratigo/pkg/migratigo"
	_ "github.com/lib/pq"
)

//go:embed templates/config_template.yml
var configTemplate string

func Initialize(dbUrl, migrationsPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	_, err = os.Stat(path.Join(cwd, "migratigo"))
	if err == nil {
		return ErrAlreadyInitialized
	}

	// Check that connection string is valid
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return err
	}
	defer func() {
		err = db.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error closing db connection:", err)
		}
	}()

	err = db.Ping()
	if err != nil {
		return err
	}

	err = migratigo.InitTable(db)
	if err != nil {
		return err
	}

	// check that migrations dir is valid
	_, err = os.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	err = os.Mkdir("migratigo", 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(cwd, "migratigo", "config.yaml"))
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error closing file (migratigo/config.yaml):", err)
		}
	}()

	filledTemplate := strings.Replace(configTemplate, "{{.DBURL}}", dbUrl, -1)
	filledTemplate = strings.Replace(filledTemplate, "{{.MigrationsPath}}", migrationsPath, -1)

	_, err = f.WriteString(filledTemplate)
	if err != nil {
		return err
	}

	return nil
}
