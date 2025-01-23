package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vvanpo/upspin-fly/dirserver/state"
	"upspin.io/upspin"
)

func (s State) List(ctx context.Context, eid state.EntryId) ([]*upspin.DirEntry, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("sqlite.List(%d): begin transaction: %w", eid, err)
	}

	rs, err := tx.Query(
		`SELECT
			e.name, e.sequence, o.timestamp, p.writer, p.dir, p.link, p.packing, p.packdata
		FROM proj_entry e
		INNER JOIN log_operation o ON e.op = o.id
		INNER JOIN log_put p ON o.put = p.id
		WHERE e.parent = ?`,
		eid,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite.List(%d): query: %w", eid, err)
	}
	defer rs.Close()

	for rs.Next() {

	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite.List(%d): commit: %w", eid, err)
	}

	return nil, nil
}

func scanEntry(rs *sql.Rows) (*upspin.DirEntry, error) {
	e := &upspin.DirEntry{}
	var name string
	var dir bool
	var link sql.NullString
	var packing sql.NullByte
	var packdata []byte
	if err := rs.Scan(&name, &e.Sequence, &e.Time, &e.Writer, &dir, &link, &packing, &packdata); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying DirEntry: %w", err)
	}

	e.Name = upspin.PathName(name)
	e.SignedName = e.Name
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
