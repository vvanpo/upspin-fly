package state

import (
	"context"
	"database/sql"

	"upspin.io/path"
	"upspin.io/upspin"
)

type State struct {
	db *sql.DB
}

// Put appends a put operation to the log.
func (s State) Put(ctx context.Context, e *upspin.DirEntry) error {
	p, _ := path.Parse(e.Name)
	tx, i, err := s.op(ctx, p)
	if err != nil {
		return err
	}

	exec := func() (err error) {
		switch e.Attr {
		case upspin.AttrDirectory:
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO log_put (operation, writer, dir) VALUES (?, ?, ?)`,
				i,
				e.Writer,
				true,
			)
		case upspin.AttrLink:
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO log_put (operation, writer, link) VALUES (?, ?, ?)`,
				i,
				e.Writer,
				e.Link,
			)
		default:
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO log_put (operation, writer, packing, packdata) VALUES (?, ?, ?, ?)`,
				i,
				e.Writer,
				e.Packing,
				e.Packdata,
			)
		}

		return err
	}

	if err = exec(); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Delete appends a delete operation to the log.
func (s State) Delete(ctx context.Context, p path.Parsed) error {
	tx, i, err := s.op(ctx, p)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO log_delete VALUES (?)`, i); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s State) root(ctx context.Context, tx *sql.Tx, name upspin.UserName) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO root (username) VALUES (?)`, name)
	return err
}

// op appends an operation to the log and returns the id.
func (s State) op(ctx context.Context, p path.Parsed) (*sql.Tx, int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, -1, err
	}

	r, err := tx.ExecContext(
		ctx,
		`INSERT INTO log_operation (root, path) VALUES ((SELECT id FROM root WHERE username = ?), ?)`,
		p.User(),
		p.FilePath(),
	)
	if err != nil {
		tx.Rollback()
		return nil, -1, err
	}

	i, err := r.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, -1, err
	}

	return tx, i, err
}
