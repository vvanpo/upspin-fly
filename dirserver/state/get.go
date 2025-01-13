package state

import (
	"context"
	"database/sql"
	"fmt"

	"upspin.io/path"
	"upspin.io/upspin"
)

// GetAll retrieves the entries for all segments in a path. If a link is found
// in the path and it is the last element returned, regardless of whether this
// completes the requested path. If an entry does not exist, an empty slice is
// returned.
func (s State) GetAll(ctx context.Context, p path.Parsed) ([]*upspin.DirEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin transaction for GetAll: %w", err)
	}

	es := make([]*upspin.DirEntry, 0, p.NElem())
	for i := 0; i <= p.NElem(); i++ {
		e, err := get(tx, p.First(i))
		if err != nil || e == nil {
			tx.Commit()
			return nil, err
		}

		es = append(es, e)

		if e.IsLink() {
			break
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing for GetAll: %w", err)
	}

	return es, nil
}

func get(tx *sql.Tx, p path.Parsed) (*upspin.DirEntry, error) {
	r := tx.QueryRow(
		`SELECT COUNT(*)
		FROM log_operation o
		INNER JOIN root r ON o.root = r.id
		LEFT JOIN log_put p ON o.put = p.id
		WHERE r.username = ?
			AND o.path = ?
		`,
		p.User(),
		p.FilePath(),
	)
	var seq int64
	if err := r.Scan(&seq); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting sequence: %w", err)
	}

	r = tx.QueryRow(
		`SELECT
			o.timestamp, p.writer, p.link, p.dir
		FROM log_operation o
		INNER JOIN root r ON o.root = r.id
		LEFT JOIN log_put p ON o.put = p.id
		WHERE r.username = ?
			AND o.path = ?
		ORDER BY o.id DESC
		LIMIT 1
		`,
		p.User(),
		p.FilePath(),
	)
	if err := r.Err(); err != nil {
		return nil, fmt.Errorf("getting DirEntry: %w", err)
	}

	e := &upspin.DirEntry{
		Name:       p.Path(),
		SignedName: p.Path(),
		Sequence:   seq,
	}
	var dir bool
	r.Scan(&e.Time, &e.Writer, &e.Link, &dir)

	if dir {
		e.Attr = upspin.AttrDirectory
	} else if e.Link != "" {
		e.Attr = upspin.AttrLink
	}

	//TODO get blocks

	return e, nil
}
