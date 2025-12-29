package migration

import "errors"

var (
	ErrInvalidMigrationName   = errors.New("invalid migration name")
	ErrInvalidMigrationNumber = errors.New("invalid migration number")
)
