package state

import (
	"context"
	"database/sql"
	"fmt"

	"upspin.io/path"
	"upspin.io/upspin"
)

// Put appends a put operation to the log.
func (s State) Put(ctx context.Context, e *upspin.DirEntry) error {
	p, _ := path.Parse(e.Name)
	tx, err := s.db.BeginTx(ctx, nil)

	if p.IsRoot() {
		if _, err := tx.Exec(`INSERT INTO root (username) VALUES (?)`, p.User()); err != nil {
			tx.Rollback()
			return fmt.Errorf("create root for %s: %w", p.User(), err)
		}
	}

	i, err := s.op(tx, p)
	if err != nil {
		return fmt.Errorf("append operation to log: %w", err)
	}

	if err = execPut(tx, i, e); err != nil {
		tx.Rollback()
		return fmt.Errorf("append put to log: %w", err)
	}

	return tx.Commit()
}

func execPut(tx *sql.Tx, op int64, e *upspin.DirEntry) (err error) {
	switch e.Attr {
	case upspin.AttrDirectory:
		_, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, dir) VALUES (?, ?, ?)`,
			op,
			e.Writer,
			true,
		)
	case upspin.AttrLink:
		_, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, link) VALUES (?, ?, ?)`,
			op,
			e.Writer,
			e.Link,
		)
	default:
		_, err = tx.Exec(
			`INSERT INTO log_put (operation, writer, packing, packdata) VALUES (?, ?, ?, ?)`,
			op,
			e.Writer,
			e.Packing,
			e.Packdata,
		)
	}

	return err
}
