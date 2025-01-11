package state

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"upspin.io/upspin"
)

func TestPutLookupDelete(t *testing.T) {
	ctx := context.Background()
	db, _ := sql.Open("sqlite3", "file:/tmp/upspin-fly/test.db?_fk=true")
	s, _ := New(db)
	if err := s.create(ctx); err != nil {
		t.Error(err)
	}
	tx, _ := db.Begin()
	s.root(ctx, tx, "foo@example.com")
	if err := tx.Commit(); err != nil {
		t.Error(err)
	}

	if err := s.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrDirectory,
		Writer: "foo@example.com",
		Name:   "foo@example.com/bar",
	}); err != nil {
		t.Error(err)
	}
	baz := &upspin.DirEntry{
		Packing: upspin.PlainPack,
		Blocks: []upspin.DirBlock{
			{
				Location: upspin.Location{
					Endpoint: upspin.Endpoint{
						Transport: upspin.Remote,
						NetAddr:   "localhost:123",
					},
					Reference: "bazref",
				},
				Offset: 0,
				Size:   24,
			},
		},
		Writer: "foo@example.com",
		Name:   "foo@example.com/bar/baz",
	}
	if err := s.Put(ctx, baz); err != nil {
		t.Error(err)
	}
	baz.Writer = "qux@example.com"
	baz.Blocks[0].Location.Reference = "bazref2"
	baz.Blocks[0].Size = 40
	if err := s.Put(ctx, baz); err != nil {
		t.Error(err)
	}
}
