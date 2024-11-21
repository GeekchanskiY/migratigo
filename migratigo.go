package migratigo

import (
	"database/sql"
	"fmt"
)

type Connector struct {
	migrated   bool
	connection *sql.DB
}

func New(host, port, username, password, name string) (*Connector, error) {
	connection, err := Connect(host, port, username, password, name)
	if err != nil {
		return nil, err
	}
	return &Connector{
		migrated:   false,
		connection: connection,
	}, nil
}

// Connect connects to database, and applies migrations
func Connect(host, port, username, password, name string) (*sql.DB, error) {
	dbInfo := createConnectionString(host, port, username, password, name)

	connection, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	err = connection.Ping()
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func ConnectFromConnectionString(connString string) (*sql.DB, error) {
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

func (c *Connector) RunMigrations() error {
	return nil
}

func (c *Connector) Close() error {
	return c.connection.Close()
}

func (c *Connector) Connection() *sql.DB {

	return c.connection
}

func createConnectionString(host, port, username, password, name string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		username,
		password,
		name,
	)
}
