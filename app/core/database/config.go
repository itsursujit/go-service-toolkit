package database

import "fmt"

const (
	// DialectMysql is the mysql dialect.
	DialectMysql = "mysql"
	// DialectPostgres is the postgres dialect.
	DialectPostgres = "postgres"
)

// Config holds all configuration values for the DB setup
type Config struct {
	Dialect  string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string // optional, only used for postgres
}

// ConnectionString returns a valid string for sql.Open.
func (c *Config) ConnectionString() string {
	switch c.Dialect {
	case DialectMysql:
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&multiStatements=true",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case DialectPostgres:
		dbName := c.Name
		if dbName == "" {
			// We probably don't have a DB yet, let's connect to `postgres` instead
			dbName = "postgres"
		}
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, dbName, c.SSLMode)
	default:
		panic("Unknown database dialect: " + c.Dialect)
	}
}

// MigrationURL returns a database URL with database for migrations.
func (c *Config) MigrationURL() string {
	switch c.Dialect {
	case DialectMysql:
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&multiStatements=true",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case DialectPostgres:
		return fmt.Sprintf("%s:%s@%s:%s/%s?sslmode=%s",
			c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
	default:
		panic("Unknown database dialect: " + c.Dialect)
	}
}

// createDatabaseQuery returns the query to create a database.
// Panics on invalid dialect.
func (c *Config) createDatabaseQuery() string {
	switch c.Dialect {
	case DialectMysql:
		return "CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci;"
	case DialectPostgres:
		return `CREATE DATABASE "%s" ENCODING=UTF8;`
	default:
		panic("Unknown database dialect: " + c.Dialect)
	}
}
