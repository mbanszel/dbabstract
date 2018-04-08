// Copyright (c) 2018 Jef Oliver. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbabstract

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	_ "github.com/go-sql-driver/mysql" // side-effects from go-sql-driver/mysql
)

var (
	mysqlTC = "SELECT table_name FROM information_schema.tables WHERE table_schema = '?' AND table_name = '?'"
)

type mysqlDBHolder struct {
	db     *sql.DB
	dbName string
	driver string
}

// Close is a wrapper function to sql.DB.Close()
func (s *mysqlDBHolder) Close() error {
	return s.db.Close()
}

// DB returns the pointer to the open SQL database connection
func (s *mysqlDBHolder) DB() *sql.DB {
	return s.db
}

// Driver returns the configured driver
func (s *mysqlDBHolder) Driver() string {
	return s.driver
}

// Format makes sure all query arguments are '?' instead of '$' or others.
func (s *mysqlDBHolder) Format(query string) string {
	return strings.Replace(query, "$", "?", -1)
}

// TableExists checks for the existence of a table in a MySQL database.
func (s *mysqlDBHolder) TableExists(table string) (bool, error) {
	var tName string

	queryStr := s.Format(mysqlTC)
	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		return false, err
	}

	err = stmt.QueryRow(s.dbName, table).Scan(&tName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if tName != table {
		return false, nil
	}

	return true, nil
}

// mysqlBuildPath builds the full path to the MySQL database
func mysqlBuildPath(opts DBOpts) string {
	var addr string

	switch opts.ConnectType {
	case TCP, TCP4, TCP6:
		addr = fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	case UNIX:
		addr = opts.SocketPath
	}

	ret := mysql.Config{
		User:   opts.Username,
		Passwd: opts.Password,
		Net:    opts.ConnectType,
		Addr:   addr,
		DBName: opts.DBName,
	}

	return ret.FormatDSN()
}

// mysqlValidateOptions validates needed options for the MySQL database driver
func mysqlValidateOptions(opts DBOpts) error {
	if len(strings.TrimSpace(opts.ConnectType)) > 0 {
		return ErrDatabaseConnectTypeMissing
	}
	if len(strings.TrimSpace(opts.DBName)) == 0 {
		return ErrDatabaseNameMissing
	}

	switch opts.ConnectType {
	case TCP, TCP4, TCP6:
		if len(strings.TrimSpace(opts.Host)) > 0 {
			return ErrDatabaseHostMissing
		}
		if opts.Port <= 0 {
			return ErrDatabasePortMissing
		}
	case UNIX:
		if len(strings.TrimSpace(opts.SocketPath)) > 0 {
			return ErrDatabaseSocketPathMissing
		}
	default:
		return ErrDatabaseConnectTypeUnsupported
	}

	return nil
}

// newDBHolderMySQL returns a connection to be used with an MySQL/MariaDB database
func newDBHolderMySQL(opts DBOpts) (DBHolder, error) {
	var err error
	var ret mysqlDBHolder

	if err := mysqlValidateOptions(opts); err != nil {
		return nil, err
	}

	dbPath := mysqlBuildPath(opts)

	ret.dbName = opts.DBName
	ret.driver = opts.Driver
	ret.db, err = sql.Open("mysql", dbPath)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func init() {
	holderFactories["mariadb"] = newDBHolderMySQL
	holderFactories["mysql"] = newDBHolderMySQL
}
