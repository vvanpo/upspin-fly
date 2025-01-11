package state

import (
	"context"
	"database/sql"
	"fmt"

	"upspin.io/path"
	"upspin.io/upspin"
)

// Put persists a put operation.
func (s State) Put(ctx context.Context, e *upspin.DirEntry) error {
	p, _ := path.Parse(e.Name)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction for Put: %w", err)
	}

	if p.IsRoot() {
		if _, err := tx.Exec(`INSERT INTO root (username) VALUES (?)`, p.User()); err != nil {
			tx.Rollback()
			return fmt.Errorf("create root for %s: %w", p.User(), err)
		}
	}

	oid, err := s.op(tx, p)
	if err != nil {
		return fmt.Errorf("persist operation to log: %w", err)
	}

	pid, err := execPut(tx, oid, e)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("persist put to log: %w", err)
	}

	for _, b := range e.Blocks {
		_, err := tx.Exec(
			`INSERT INTO block VALUES (?, ?, ?, ?, ?, ?)`,
			pid,
			b.Location.Endpoint.NetAddr,
			b.Location.Reference,
			b.Offset,
			b.Size,
			b.Packdata,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("persist block %v: %w", b.Location, err)
		}
	}

	return tx.Commit()
}

func execPut(tx *sql.Tx, op int64, e *upspin.DirEntry) (int64, error) {
	var r sql.Result
	var err error
	switch e.Attr {
	case upspin.AttrDirectory:
		r, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, dir) VALUES (?, ?, ?)`,
			op,
			e.Writer,
			true,
		)
	case upspin.AttrLink:
		r, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, link) VALUES (?, ?, ?)`,
			op,
			e.Writer,
			e.Link,
		)
	default:
		r, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, packing, packdata) VALUES (?, ?, ?, ?)`,
			op,
			e.Writer,
			e.Packing,
			e.Packdata,
		)
	}

	i, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return i, err
}
