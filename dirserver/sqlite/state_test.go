package sqlite

import (
	"context"
	"testing"

	"upspin.io/path"
	"upspin.io/upspin"
)

func TestPutGetDelete(t *testing.T) {
	ctx := context.Background()
	s, err := Open("/tmp/upspin-fly/test.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	/// Puts
	if err := s.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrDirectory,
		Writer: "foo@example.com",
		Name:   "foo@example.com/",
	}); err != nil {
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
		Writer:     "foo@example.com",
		Name:       "foo@example.com/bar/baz",
		SignedName: "foo@example.com/bar/baz",
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

	bazp, _ := path.Parse(baz.Name)

	/// Get
	es, err := s.GetAll(ctx, bazp)
	if err != nil {
		t.Error(err)
	}
	if len(es) != 3 {
		t.Fatalf("wrong number of entries: %d", len(es))
	}
	e := es[len(es)-1]
	if e.Name != "foo@example.com/bar/baz" {
		t.Errorf("wrong name for baz: %s", e.Name)
	}
	if es[1].Attr != upspin.AttrDirectory {
		t.Errorf("bar not a directory: %v", es[1])
	}
	if e.Sequence != 4 {
		t.Errorf("wrong sequence for baz: %d", e.Sequence)
	}
	if es[0].Sequence != 4 {
		t.Errorf("wrong sequence for root: %d", es[0].Sequence)
	}

	/// Delete
	if err := s.Delete(ctx, bazp); err != nil {
		t.Error(err)
	}
}
