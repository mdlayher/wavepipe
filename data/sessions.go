package data

import (
	"database/sql"

	"github.com/mdlayher/wavepipe/data/models"
)

const (
	// sqlSelectSessionByKey is the SQL statement used to select a single Session
	// by key
	sqlSelectSessionByKey = `
		SELECT * FROM sessions WHERE key = ?;
	`

	// sqlInsertSession is the SQL statement used to insert a new Session
	sqlInsertSession = `
		INSERT INTO sessions (
			"user_id"
			, "key"
			, "expire"
			, "client"
		) VALUES (?, ?, ?, ?);
	`

	// sqlUpdateSession is the SQL statement used to update an existing Session
	sqlUpdateSession = `
		UPDATE sessions SET
			"user_id" = ?
			, "key" = ?
			, "expire" = ?
			, "client" = ?
		WHERE id = ?;
	`

	// sqlDeleteSession is the SQL statement used to delete an existing Session
	sqlDeleteSession = `
		DELETE FROM sessions WHERE id = ?;
	`

	// sqlDeleteSessionsByUserID is the SQL statement used to delete all Sessions
	// for a user, by the user's ID
	sqlDeleteSessionsByUserID = `
		DELETE FROM sessions WHERE user_id = ?;
	`
)

// SelectSessionByKey returns a single Session by key from the database.
func (db *DB) SelectSessionByKey(key string) (*models.Session, error) {
	return db.selectSingleSession(sqlSelectSessionByKey, key)
}

// InsertSession starts a transaction, inserts a new Session, and attempts to commit
// the transaction.
func (db *DB) InsertSession(s *models.Session) error {
	return db.WithTx(func(tx *Tx) error {
		return tx.InsertSession(s)
	})
}

// UpdateSession starts a transaction, updates the input Session by its ID, and attempts
// to commit the transaction.
func (db *DB) UpdateSession(s *models.Session) error {
	return db.WithTx(func(tx *Tx) error {
		return tx.UpdateSession(s)
	})
}

// DeleteSession starts a transaction, deletes the input Session by its ID, and attempts
// to commit the transaction.
func (db *DB) DeleteSession(s *models.Session) error {
	return db.WithTx(func(tx *Tx) error {
		return tx.DeleteSession(s)
	})
}

// DeleteSessionsByUserID starts a transaction, deletes all Sessions with the matching user ID,
// and attempts to commit the transaction.
func (db *DB) DeleteSessionsByUserID(userID uint64) error {
	return db.WithTx(func(tx *Tx) error {
		return tx.DeleteSessionsByUserID(userID)
	})
}

// selectSessions returns a slice of Sessions from the database, based upon an input
// SQL query and arguments
func (db *DB) selectSessions(query string, args ...interface{}) ([]*models.Session, error) {
	// Slice of sessions to return
	var sessions []*models.Session

	// Invoke closure with prepared statement and wrapped rows,
	// passing any arguments from the caller
	err := db.withPreparedRows(query, func(rows *Rows) error {
		// Scan rows into a slice of Sessions
		var err error
		sessions, err = rows.ScanSessions()

		// Return errors from scanning
		return err
	}, args...)

	// Return any matching sessions and error
	return sessions, err
}

// selectSingleSession returns a Session from the database, based upon an input
// SQL query and arguments
func (db *DB) selectSingleSession(query string, args ...interface{}) (*models.Session, error) {
	// Fetch sessions with matching condition
	sessions, err := db.selectSessions(query, args...)
	if err != nil {
		return nil, err
	}

	// Verify only 0 or 1 session returned
	if len(sessions) == 0 {
		return nil, sql.ErrNoRows
	} else if len(sessions) == 1 {
		return sessions[0], nil
	}

	// More than one result returned
	return nil, ErrMultipleResults
}

// InsertSession inserts a new Session in the context of the current transaction.
func (tx *Tx) InsertSession(s *models.Session) error {
	// Execute SQL to insert Session
	result, err := tx.Tx.Exec(sqlInsertSession, s.SQLWriteFields()...)
	if err != nil {
		return err
	}

	// Retrieve generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Store generated ID
	s.ID = uint64(id)
	return nil
}

// UpdateSession updates the input Session by its ID, in the context of the
// current transaction.
func (tx *Tx) UpdateSession(s *models.Session) error {
	_, err := tx.Tx.Exec(sqlUpdateSession, s.SQLWriteFields()...)
	return err
}

// DeleteSession updates the input Session by its ID, in the context of the
// current transaction.
func (tx *Tx) DeleteSession(s *models.Session) error {
	_, err := tx.Tx.Exec(sqlDeleteSession, s.ID)
	return err
}

// DeleteSessionsByUserID deletes all Sessions with the input user ID, in the
// context of the current transaction.
func (tx *Tx) DeleteSessionsByUserID(userID uint64) error {
	_, err := tx.Tx.Exec(sqlDeleteSessionsByUserID, userID)
	return err
}

// ScanSessions returns a slice of Sessions from wrapped rows.
func (r *Rows) ScanSessions() ([]*models.Session, error) {
	// Iterate all returned rows
	var sessions []*models.Session
	for r.Rows.Next() {
		// Scan new session into struct, using specified fields
		s := new(models.Session)
		if err := r.Rows.Scan(s.SQLReadFields()...); err != nil {
			return nil, err
		}

		// Discard any nil results
		if s == nil {
			continue
		}

		// Append session to output slice
		sessions = append(sessions, s)
	}

	return sessions, nil
}
