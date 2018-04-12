package dbabstract

import (
	"database/sql"
	"strings"
)

type commonDB struct {
	db     *sql.DB
	dbName string
	driver string
	path   string
}

// Close is a wrapper function to sql.DB.Close()
func (c *commonDB) Close() error {
	return c.db.Close()
}

// DB returns the pointer to the open SQL database connection
func (c *commonDB) DB() *sql.DB {
	return c.db
}

// Driver returns the configured driver
func (c *commonDB) Driver() string {
	return c.driver
}

// Format makes sure all query arguments are '?' instead of '$' or others.
func (c *commonDB) Format(query string) string {
	return strings.Replace(query, "$", "?", -1)
}

// Path returns the path used to connect to the database. This is useful for debug purposes.
// the username and password are not stored.
func (c *commonDB) Path() string {
	return c.path
}
