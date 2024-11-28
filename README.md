# migratigo

Migratigo is a yet another lightweight migration tool.

## Usage

To install:

```shell
go get github.com/GeekchanskiY/migratigo
go mod tidy
```

Example usage:

```go
package main

import (
	"embed"
	
	"github.com/GeekchanskiY/migratigo"
)

const (
	migrationsDir = "migrations" // name of the directory where embed migrations located
	connString = "..."
)

//go:embed migrations/*.sql
var migrations embed.FS

func main(){
	connection, err := migratigo.Connect(connString)
	if err != nil {
		// ...
    }
	connector, err := migratigo.New(connection, migrations, migrationsDir)
	if err != nil {
		// ...
    }
	err = connector.RunMigrations() 
	if err != nil {
		// ...
    }
	// migrations are saved and applied, use your database :)
}
```

## Future plans
CLI tools
Downgrade migrations
Auto-create database
ORM features
Implement database/sql connection interface for connector, maybe
