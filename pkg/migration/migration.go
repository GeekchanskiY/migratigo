package migration

import (
	"os"
	"path/filepath"
)

type Migration struct {
	Num      int
	Title    string
	Up       bool
	Migrated bool
	Content  string
}

func FromFile(filePath string) (*Migration, error) {
	err := validateMigrationName(filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	num, title, up, err := formatName(filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &Migration{
		Num:      num,
		Title:    title,
		Up:       up,
		Migrated: false,
		Content:  string(content),
	}, nil
}
