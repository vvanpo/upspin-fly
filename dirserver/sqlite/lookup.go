package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"upspin.io/path"
	"upspin.io/upspin"
)

// LookupAll implements dirserver.State.
func (s State) LookupAll(ctx context.Context, p path.Parsed) ([]*upspin.DirEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin transaction for LookupAll: %w", err)
	}

	es := make([]*upspin.DirEntry, 0, p.NElem())
	for i := 0; i <= p.NElem(); i++ {
		e, err := get(tx, p.First(i))
		if err != nil {
			tx.Commit()
			return nil, err
		} else if e == nil {
			break
		}

		es = append(es, e)

		if e.IsLink() {
			break
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing for LookupAll: %w", err)
	}

	return es, nil
}

func get(tx *sql.Tx, p path.Parsed) (*upspin.DirEntry, error) {
	r := tx.QueryRow(
		`SELECT
			e.sequence, o.timestamp, p.writer, p.dir, p.link, p.packing, p.packdata
		FROM cache_entry e
		INNER JOIN log_operation o ON e.op = o.id
		INNER JOIN root r ON o.root = r.id
		INNER JOIN log_put p ON o.put = p.id
		WHERE e.name = ?
		`,
		p.Path(),
	)
	if err := r.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting DirEntry: %w", err)
	}

	e := &upspin.DirEntry{
		Name:       p.Path(),
		SignedName: p.Path(),
	}
	var dir bool
	var link sql.NullString
	var packing sql.NullByte
	var packdata []byte
	if err := r.Scan(&e.Sequence, &e.Time, &e.Writer, &dir, &link, &packing, &packdata); err != nil {
		return nil, err
	}
	if dir {
		e.Attr = upspin.AttrDirectory
	} else if link.Valid {
		e.Attr = upspin.AttrLink
		e.Link = upspin.PathName(link.String)
	} else {
		e.Attr = upspin.AttrIncomplete
		e.Packing = upspin.Packing(packing.Byte)
		if len(packdata) > 0 {
			e.Packdata = packdata
		}
	}

	return e, nil
}
