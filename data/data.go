// Package data provides the database abstraction layer and helpers
// for the wavepipe media server.
package data

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/mattn/go-sqlite3"
)

const (
	// driverSqlite3 is the name of the sqlite3 database/sql driver.
	driverSqlite3 = "sqlite3"
)

var (
	// ErrMultipleResults is returned when a query should return only zero or a single
	// result, but returns two or more results.
	ErrMultipleResults = errors.New("db: multiple results returned")
)

// DB provides the database abstraction layer for the application.
type DB struct {
	*sql.DB

	driver string

	// preparedStmts is a map of query strings to prepared database statements.
	// On first use, queries are prepared and added to the map for later re-use.
	// On shutdown, all prepared statementes are cleaned up.
	preparedStmts map[string]*sql.Stmt
	stmtMutex     *sync.RWMutex
}

// Open opens and initializes a database instance.
func (db *DB) Open(driver string, dsn string) error {
	// Open database
	d, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	db.DB = d
	db.driver = driver

	// Perform driver-specific setup
	if db.driver == driverSqlite3 {
		if err := db.sqlite3Setup(); err != nil {
			return err
		}
	}

	// Initialize prepared statement map and mutex
	db.preparedStmts = make(map[string]*sql.Stmt)
	db.stmtMutex = new(sync.RWMutex)

	return nil
}

// Close closes and cleans up a database instance.
func (db *DB) Close() error {
	// Clean up all prepared statements
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()
	for _, v := range db.preparedStmts {
		if err := v.Close(); err != nil {
			return err
		}
	}

	return db.DB.Close()
}

// Begin starts a transaction on this database instance.
func (db *DB) Begin() (*Tx, error) {
	// Start a transaction on underlying database
	dbtx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	// Return wrapped transaction
	return &Tx{
		Tx: dbtx,
	}, nil
}

// IsConstraintFailure returns whether or not an input error is due to a failed
// database constraint, such as an insert of an item which is not unique.
func (db *DB) IsConstraintFailure(err error) bool {
	// sqlite3-specific constraint failure checking
	if db.driver == driverSqlite3 {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			return sqliteErr.Code == sqlite3.ErrConstraint
		}
	}

	// Not a constraint failure
	return false
}

// IsReadonly returns whether or not an input error is due to a readonly database.
func (db *DB) IsReadonly(err error) bool {
	// sqlite3-specific readonly checking
	if db.driver == driverSqlite3 {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			return sqliteErr.Code == sqlite3.ErrReadonly
		}
	}

	// Not readonly
	return false
}

// WithTx creates a new wrapped transaction, invokes an input closure, and
// commits or rolls back the transaction, depending on the result of the
// closure invocation.
func (db *DB) WithTx(fn func(tx *Tx) error) error {
	// Start a wrapped transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Invoke input closure, passing in wrapped transaction
	if err := fn(tx); err != nil {
		// Failure, attempt to roll back transaction
		if rErr := tx.Rollback(); rErr != nil {
			return rErr
		}

		// Return error from closure
		return err
	}

	// Attempt to commit transaction
	return tx.Commit()
}

// withPreparedStmt creates or re-uses a prepared statement for the input SQL query.
// On first use, prepared statements are created and set into the preparedStmts map
// for later reuse.
func (db *DB) withPreparedStmt(query string, fn func(stmt *sql.Stmt) error) error {
	// Check for pre-existing statement
	db.stmtMutex.RLock()
	stmt, ok := db.preparedStmts[query]
	db.stmtMutex.RUnlock()
	if !ok {
		// Prepare statement using input query
		var err error
		stmt, err = db.Prepare(query)
		if err != nil {
			return err
		}

		// Store statement for reuse
		db.stmtMutex.Lock()
		db.preparedStmts[query] = stmt
		db.stmtMutex.Unlock()
	}

	// Invoke input closure with prepared statement, return results of closure
	return fn(stmt)
}

// withPreparedRows creates or retrieves a prepared statement with the input SQL query,
// invokes an input closure containing SQL rows, and handles cleanup of rows
// once the closure invocation is complete.
func (db *DB) withPreparedRows(query string, fn func(rows *Rows) error, args ...interface{}) error {
	// Create or retrieve a prepared statement
	return db.withPreparedStmt(query, func(stmt *sql.Stmt) error {
		// Perform input query, sending arguments from caller
		rows, err := stmt.Query(args...)
		if err != nil {
			return err
		}

		// Invoke input closure with wrapped Rows type, capturing return value for later
		fnErr := fn(&Rows{
			Rows: rows,
		})

		// Close rows
		if err := rows.Close(); err != nil {
			return err
		}

		// Check errors from rows
		if err := rows.Err(); err != nil {
			return err
		}

		// Return result of closure
		return fnErr
	})
}

// sqlite3Setup performs setup routines specific to the sqlite3 database driver,
// each time the database is initialized.
func (db *DB) sqlite3Setup() error {
	// Slice of queries which will be executed, in-order, on startup
	queries := []string{
		// Enforce foreign keys
		"PRAGMA foreign_keys = ON;",

		// Do not wait for data to be written to disk
		"PRAGMA synchronous = OFF;",

		// Keep rollback journal in memory
		"PRAGMA journal_mode = MEMORY;",
	}

	// Execute startup queries in order
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// Tx is a wrapped database transaction, which provides additional methods
// for interacting directly with custom types.
type Tx struct {
	*sql.Tx
}

// Rows is a wrapped set of database rows, which provides additional methods
// for interacting directly with custom types.
type Rows struct {
	*sql.Rows
}
