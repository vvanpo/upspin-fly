package dirserver

import (
	"context"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/upspin"
)

func TestWhichAccess(t *testing.T) {
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
		Name:    "foo@example.com/Access",
	})

	s := server{State: st}
	e, err := s.WhichAccess("foo@example.com/bar")
	if err != nil {
		t.Error(err)
	} else if e == nil {
		t.Errorf("no access file found")
	} else if e.Name != "foo@example.com/Access" {
		t.Errorf("access DirEntry has wrong name: %s", e.Name)
	}
}
