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

package migration

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	validMigrationNameRegex = `^\d{3}_[a-zA-Z0-9_]+(?:\.up|\.down)\.sql$`
	getMigrationDetailRegex = `^(\d{3})_([a-zA-Z0-9_]+)\.(up|down)\.sql$`
)

func validateMigrationName(name string) error {
	regex := regexp.MustCompile(validMigrationNameRegex)

	if !regex.MatchString(name) {
		return fmt.Errorf("migration name '%s' is not valid", name)
	}

	return nil
}

func formatName(filename string) (num int, title string, up bool, err error) {
	regex := regexp.MustCompile(getMigrationDetailRegex)
	matches := regex.FindStringSubmatch(filename)

	// additional check, if validateMigrationName fails
	if len(matches) != 4 {
		return 0, "", false, ErrInvalidMigrationName
	}

	// get all args from Migration name
	num, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", false, err
	}

	if num <= 0 || num > 999 {
		return 0, "", false, ErrInvalidMigrationNumber
	}

	title = matches[2]

	if matches[3] == "up" {
		up = true
	} else {
		up = false
	}

	return
}
