package dirserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/upspin"
)

func TestGlob(t *testing.T) {
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
