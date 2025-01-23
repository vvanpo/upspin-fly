package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vvanpo/upspin-fly/dirserver/state"
	"upspin.io/path"
	"upspin.io/upspin"
)

func (s State) LookupElem(ctx context.Context, p path.Parsed) (state.EntryId, path.Parsed, upspin.Attribute, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return -1, path.Parsed{}, 0, fmt.Errorf("begin transaction for LookupElem(%s): %w", p, err)
	}

	eid := state.EntryId(-1)
	var ep path.Parsed
	var attr upspin.Attribute
	for i := 0; i <= p.NElem(); i++ {
		id, a, err := getAttr(tx, p.First(i).Path())
		if err != nil {
			tx.Commit()
			return -1, path.Parsed{}, 0, err
		} else if id == -1 {
			break
		}

		eid = id
		ep = p.First(i)
		attr = a
		if a != upspin.AttrDirectory {
			// Only continue with lookups if we know there might be a child
			// element.
			break
		}
	}

	if err := tx.Commit(); err != nil {
		return -1, path.Parsed{}, 0, fmt.Errorf("committing for LookupElem(%s): %w", p, err)
	}

	return eid, ep, attr, nil
}

// LookupAll implements state.State.
func (s State) LookupAll(ctx context.Context, p path.Parsed) ([]*upspin.DirEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin transaction for LookupAll(%s): %w", p, err)
	}

	es := make([]*upspin.DirEntry, 0, p.NElem())
	for i := 0; i <= p.NElem(); i++ {
		e, err := get(tx, p.First(i).Path())
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
		return nil, fmt.Errorf("committing for LookupAll(%s): %w", p, err)
	}

	return es, nil
}

// Lookup implements state.State.
func (s State) Lookup(ctx context.Context, name upspin.PathName) (*upspin.DirEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin transaction for Lookup(%s): %w", name, err)
	}

	e, err := get(tx, name)
	if err != nil {
		tx.Commit()
		return nil, err
	}
	// TODO get blocks

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing for Lookup(%s): %w", name, err)
	}

	return e, nil
}

func get(tx *sql.Tx, name upspin.PathName) (*upspin.DirEntry, error) {
	r := tx.QueryRow(
		`SELECT
			e.sequence, o.timestamp, p.writer, p.dir, p.link, p.packing, p.packdata
		FROM proj_entry e
		INNER JOIN log_operation o ON e.op = o.id
		INNER JOIN log_root r ON o.root = r.id
		INNER JOIN log_put p ON o.put = p.id
		WHERE e.name = ?`,
		name,
	)

	e := &upspin.DirEntry{
		Name:       name,
		SignedName: name,
	}
	var dir bool
	var link sql.NullString
	var packing sql.NullByte
	var packdata []byte
	if err := r.Scan(&e.Sequence, &e.Time, &e.Writer, &dir, &link, &packing, &packdata); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying DirEntry: %w", err)
	}
	if dir {
		e.Attr = upspin.AttrDirectory
	} else if link.Valid {
		e.Attr = upspin.AttrLink
		e.Link = upspin.PathName(link.String)
	} else {
		e.Packing = upspin.Packing(packing.Byte)
		e.Packdata = packdata
	}

	return e, nil
}

func getAttr(tx *sql.Tx, name upspin.PathName) (state.EntryId, upspin.Attribute, error) {
	r := tx.QueryRow(
		`SELECT p.id, p.dir, p.link
		FROM proj_entry e
		INNER JOIN log_operation o ON e.op = o.id
		INNER JOIN log_put p ON o.put = p.id
		WHERE e.name = ?`,
		name,
	)

	var eid state.EntryId
	var dir bool
	var link sql.NullString
	if err := r.Scan(&eid, &dir, &link); err != nil {
		if err == sql.ErrNoRows {
			return -1, 0, nil
		}
		return -1, 0, fmt.Errorf("querying entry: %w", err)
	}

	attr := upspin.AttrNone
	if dir {
		attr = upspin.AttrDirectory
	} else if link.Valid {
		attr = upspin.AttrLink
	}

	return eid, attr, nil
}

func (s State) getEntry(eid state.EntryId) (*upspin.DirEntry, error) {
	return nil, nil
}
