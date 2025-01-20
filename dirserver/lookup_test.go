package dirserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/upspin"
)

func TestLookup(t *testing.T) {
	st, _ := sqlite.Open(":memory:")
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
		Name:    "foo@example.com/bar",
	})

	s := &server{state: st}
	d := &dialed{s, slog.Default(), "foo@example.com"}

	e, err := d.Lookup("foo@example.com/bar")
	if err != nil {
		t.Error(err)
	} else if e.Name != "foo@example.com/bar" {
		t.Errorf("DirEntry has wrong name: %s", e.Name)
	}
}

func TestErrFollowLink(t *testing.T) {
	st, _ := sqlite.Open(":memory:")
	defer st.Close()
	ctx := context.Background()
	st.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrDirectory,
		Writer: "foo@example.com",
		Name:   "foo@example.com/",
	})
	st.Put(ctx, &upspin.DirEntry{
		Attr:   upspin.AttrLink,
		Link:   "foo@example.com/baz",
		Writer: "foo@example.com",
		Name:   "foo@example.com/bar",
	})

	s := &server{state: st}
	d := &dialed{s, slog.Default(), "foo@example.com"}
	e, err := d.Lookup("foo@example.com/bar/baz")
	if err == nil {
		t.Errorf("ErrFollowLink not returned: %v", e)
	}
	if err != upspin.ErrFollowLink {
		t.Error(err)
	}
}
