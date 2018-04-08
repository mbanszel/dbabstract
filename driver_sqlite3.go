// Copyright (c) 2018 Jef Oliver. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbabstract

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3" // side-effects from go-sqlite3
)

var sqlite3TC = "SELECT name FROM sqlite_master WHERE type='table' AND name='?'"

type sqlite3DBHolder struct {
	db     *sql.DB
	driver string
}

// Close is a wrapper function to sql.DB.Close()
func (s *sqlite3DBHolder) Close() error {
	return s.db.Close()
}

// DB returns the pointer to the open SQL database connection
func (s *sqlite3DBHolder) DB() *sql.DB {
	return s.db
}

// Driver returns the configured driver
func (s *sqlite3DBHolder) Driver() string {
	return s.driver
}

// Format makes sure all query arguments are '?' instead of '$' or others.
func (s *sqlite3DBHolder) Format(query string) string {
	return strings.Replace(query, "$", "?", -1)
}

// TableExists checks for the existence of a table in an sqlite3 database.
func (s *sqlite3DBHolder) TableExists(table string) (bool, error) {
	var tName string

	queryStr := s.Format(sqlite3TC)
	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		return false, err
	}

	err = stmt.QueryRow(table).Scan(&tName)
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

// sqlite3BuildPath builds the full path to the sqlite3 database
func sqlite3BuildPath(opts DBOpts) (string, error) {
	pData, err := os.Stat(opts.DataDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
			return "", err
		}
	}

	if !pData.IsDir() {
		return "", ErrDataPathNotDir
	}

	return filepath.Join(opts.DataDir, fmt.Sprintf("%s.db", opts.DBName)), nil
}

// validateOptions validates needed options for the sqlite3 database driver
func sqlite3validateOptions(opts DBOpts) error {
	if len(strings.TrimSpace(opts.DataDir)) == 0 {
		return ErrDataPathMissing
	}
	if len(strings.TrimSpace(opts.DBName)) == 0 {
		return ErrDatabaseNameMissing
	}

	return nil
}

// newDBSqlite3 returns a connection to be used with an sqlite3 database
func newDBHolderSqlite3(opts DBOpts) (DBHolder, error) {
	var ret sqlite3DBHolder

	if err := sqlite3validateOptions(opts); err != nil {
		return nil, err
	}

	dbFilePath, err := sqlite3BuildPath(opts)
	if err != nil {
		return nil, err
	}

	ret.driver = opts.Driver
	ret.db, err = sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func init() {
	holderFactories["sqlite3"] = newDBHolderSqlite3
}
