package sqlite

import (
	"context"
	"testing"

	"upspin.io/path"
	"upspin.io/upspin"
)

func TestPutLookupAll(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
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
		Packing:  upspin.PlainPack,
		Packdata: []byte("packd"),
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

	/// Lookup
	es, err := s.LookupAll(ctx, bazp)
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
	if e.Name != e.SignedName {
		t.Errorf("signed name not equal to name: %s", e.SignedName)
	}
	if e.Blocks != nil {
		t.Errorf("baz contains blocks: %v", e.Blocks)
	}
	if e.Packing != upspin.PlainPack {
		t.Errorf("incorrect packing for baz: %x", e.Packing)
	}
	if string(e.Packdata) != "packd" {
		t.Errorf("incorrect packdata for baz: %v", e.Packdata)
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
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
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
	bar := &upspin.DirEntry{
		Writer: "foo@example.com",
		Name:   "foo@example.com/bar",
	}
	if err := s.Put(ctx, bar); err != nil {
		t.Error(err)
	}

	/// Delete
	barp, _ := path.Parse(bar.Name)
	if err := s.Delete(ctx, barp); err != nil {
		t.Error(err)
	}

	/// LookupAll doesn't return deleted
	es, err := s.LookupAll(ctx, barp)
	if err != nil {
		t.Error(err)
	}
	if len(es) != 1 {
		t.Errorf("Entry not deleted: %v", es)
	}
	if es[0].Sequence != 3 {
		t.Errorf("Root sequence not incremented on delete: %d", es[0].Sequence)
	}
}

// LookupElem returns a negative id and no error when the root for the
// requested does not exist.
func TestLookupElemNoRoot(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	p, _ := path.Parse("a@b.com/foo/bar")
	eid, rp, _, err := s.LookupElem(ctx, p)
	if err != nil {
		t.Error(err)
	}
	if eid >= 0 {
		t.Errorf("non-negative entry identifier: %d", eid)
	}
	if rp.NElem() > 0 {
		t.Errorf("wrong number of elements: %d", rp.NElem())
	}
}

// LookupElem returns the root path when no other elements of the rquested path
// match.
func TestLookupElemOnlyRoot(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	if err := s.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrDirectory,
		Writer: "a@b.com",
		Name:   "a@b.com/",
	}); err != nil {
		t.Error(err)
	}

	p, _ := path.Parse("a@b.com/foo/bar")
	eid, rp, _, err := s.LookupElem(ctx, p)
	if err != nil {
		t.Error(err)
	}
	if eid < 0 {
		t.Errorf("invalid entry identifier: %d", eid)
	}
	if rp.String() != "a@b.com/" {
		t.Errorf("path doesn't match root: %s", rp.String())
	}
}

// LookupElem returns a partial path that prefixes the requested path when the
// requested path does not exist.

// LookupElem returns AttrNone when the returned entry is a file.

// LookupElem returns AttrDirectory when the returned entry is a directory.

// LookupElem returns AttrLink when the returned entry is a link.
