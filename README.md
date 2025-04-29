# migratigo

Migratigo is yet another lightweight migration tool.

## Usage

Installation:

```shell
go get github.com/GeekchanskiY/migratigo
```


To install additional cli utilities
```shell
go install github.com/GeekchanskiY/migratigo
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
	err = connector.RunMigrations(false) 
	if err != nil {
		// ...
    }
	// migrations are saved and applied, use your database :)
}
```

## Future plans
 - Downgrade migrations
 - CLI
 - Auto-create database
 - Add different database support
 - ORM features
 - Implement database/sql connection interface for connector, maybe
