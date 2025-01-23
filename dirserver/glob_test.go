package dirserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/upspin"
)

func TestGlob(t *testing.T) {
	st, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	st.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrDirectory,
		Writer: "foo@example.com",
		Name:   "foo@example.com/",
	})
	st.Put(ctx, &upspin.DirEntry{
		Packing: upspin.PlainPack,
		Writer:  "foo@example.com",
		Name:    "foo@example.com/foo",
	})
	st.Put(ctx, &upspin.DirEntry{
		Packing: upspin.PlainPack,
		Writer:  "foo@example.com",
		Name:    "foo@example.com/bar",
	})
	st.Put(ctx, &upspin.DirEntry{
		Packing: upspin.PlainPack,
		Writer:  "foo@example.com",
		Name:    "foo@example.com/baz",
	})

	s := &server{state: st, cache: &cache{}}
	d := &dialed{s, slog.Default(), "foo@example.com"}

	es, err := d.Glob("foo@example.com/bar")
	if err != nil {
		t.Error(err)
	} else if len(es) != 1 {
		t.Errorf("wrong number of DirEntrys: %d", len(es))
	} else if es[0].Name != "foo@example.com/bar" {
		t.Errorf("DirEntry has wrong name: %s", es[0].Name)
	}

	es, err = d.Glob("foo@example.com/ba*")
	if err != nil {
		t.Error(err)
	} else if len(es) != 2 {
		t.Errorf("wrong number of DirEntrys: %d", len(es))
	} else if es[0].Name != "foo@example.com/bar" {
		t.Errorf("first DirEntry has wrong name: %s", es[0].Name)
	} else if es[1].Name != "foo@example.com/baz" {
		t.Errorf("second DirEntry has wrong name: %s", es[1].Name)
	}
}

// If
// - the pattern is a path without metacharacters, and
// - the requester has list and read rights, and
// - the path resolves to a regular file, then
// - return the complete file entry

// If
// - the pattern is a path without metacharacters, and
// - the requester has list but no read rights, and
// - the path resolves to a regular file, then
// - return the file entry marked as incomplete

// If
// - the pattern is a path without metacharacters, and
// - the requester has read but no list rights, and
// - the path resolves to a regular file, then
// - return the complete file entry

// If
// - the pattern is a path without metacharacters, and
// - the requester has no rights, then
// - return access.Private error

// If
// - the pattern is a path without metacharacters, and
// - the requester has a non-read/list right (like create), and
// - the path resolves to a link entry, then
// - return the link entry without error

// If
// - the pattern a metacharacter as the last element, and
// - the pattern without the last element resolves to a regular file, then
// - return no entries and no error
