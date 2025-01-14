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

	if _, err := s.appendOp(tx, p, -1); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// appendOp appends an operation to the log and returns its id. pid is the id
// of the corresponding log_put record, <0 indicates a deletion operation.
func (s State) appendOp(tx *sql.Tx, p path.Parsed, pid int64) (int64, error) {
	var r sql.Result
	var err error
	if pid < 0 {
		r, err = tx.Exec(
			`INSERT INTO log_operation (root, path) VALUES ((SELECT id FROM root WHERE username = ?), ?)`,
			p.User(),
			p.FilePath(),
		)
	} else {
		r, err = tx.Exec(
			`INSERT INTO log_operation (root, path, put) VALUES ((SELECT id FROM root WHERE username = ?), ?, ?)`,
			p.User(),
			p.FilePath(),
			pid,
		)
	}
	if err != nil {
		tx.Rollback()
	}

	i, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return i, err
}
