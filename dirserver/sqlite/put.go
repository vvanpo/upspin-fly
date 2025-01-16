package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"upspin.io/path"
	"upspin.io/upspin"
)

// Put implements dirserver.State.
func (s State) Put(ctx context.Context, e *upspin.DirEntry) error {
	p, _ := path.Parse(e.Name)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction for Put: %w", err)
	}

	if p.IsRoot() {
		if _, err := tx.Exec(`INSERT INTO log_root (username) VALUES (?)`, p.User()); err != nil {
			tx.Rollback()
			return fmt.Errorf("create root for %s: %w", p.User(), err)
		}
	}

	pid, err := appendPut(tx, e)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("persist put to log: %w", err)
	}

	oid, err := s.appendOp(tx, p, pid)
	if err != nil {
		return fmt.Errorf("persist operation to log: %w", err)
	}

	for _, b := range e.Blocks {
		_, err := tx.Exec(
			`INSERT INTO log_block VALUES (?, ?, ?, ?, ?, ?)`,
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

	if err := projPut(tx, p, oid); err != nil {
		tx.Rollback()
		return fmt.Errorf("caching put: %w", err)
	}
	return tx.Commit()
}

func appendPut(tx *sql.Tx, e *upspin.DirEntry) (int64, error) {
	var r sql.Result
	var err error
	switch e.Attr {
	case upspin.AttrDirectory:
		r, err = tx.Exec(
			`INSERT INTO log_put (writer, dir) VALUES (?, ?)`,
			e.Writer,
			true,
		)
	case upspin.AttrLink:
		r, err = tx.Exec(
			`INSERT INTO log_put (writer, link) VALUES (?, ?)`,
			e.Writer,
			e.Link,
		)
	default:
		r, err = tx.Exec(
			`INSERT INTO log_put (writer, packing, packdata) VALUES (?, ?, ?)`,
			e.Writer,
			e.Packing,
			e.Packdata,
		)
	}
	if err != nil {
		return -1, err
	}

	i, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return i, err
}
