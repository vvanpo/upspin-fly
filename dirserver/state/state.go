package state

import (
	"context"
	"database/sql"

	"upspin.io/path"
)

type State struct {
	db *sql.DB
}

// Delete appends a delete operation to the log.
func (s State) Delete(ctx context.Context, p path.Parsed) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	i, err := s.op(tx, p)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(`INSERT INTO log_delete VALUES (?)`, i); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// op appends an operation to the log and returns the id.
func (s State) op(tx *sql.Tx, p path.Parsed) (int64, error) {
	r, err := tx.Exec(
		`INSERT INTO log_operation (root, path) VALUES ((SELECT id FROM root WHERE username = ?), ?)`,
		p.User(),
		p.FilePath(),
	)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	i, err := r.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	return i, err
}
