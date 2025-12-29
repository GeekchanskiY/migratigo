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
