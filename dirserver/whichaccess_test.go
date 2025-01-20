package dirserver

import (
	"context"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/access"
	"upspin.io/upspin"
)

type nilCache struct{}

func (_ nilCache) GetAccess(ctx context.Context, e *upspin.DirEntry) (*access.Access, error) {
	return nil, nil
}
func (_ nilCache) GetGroup(ctx context.Context, n upspin.PathName) ([]byte, error) {
	return nil, nil
}
func (_ nilCache) RemoveGroup(ctx context.Context, n upspin.PathName) error {
	return nil
}

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
	st.Put(ctx, &upspin.DirEntry{
		Packing: upspin.PlainPack,
		Writer:  "foo@example.com",
		Name:    "foo@example.com/bar",
	})

	s := &server{state: st, cache: nilCache{}}
	d := &dialed{s, "user@example.com"}

	assertAcc := func(expect upspin.PathName, e *upspin.DirEntry, err error) {
		if err != nil {
			t.Error(err)
		} else if e == nil {
			t.Errorf("no access file found")
		} else if e.Name != expect {
			t.Errorf("access DirEntry has wrong name: %s (expected %s)", e.Name, expect)
		}
	}

	// A regular file should return an adjacent access file
	e, err := d.WhichAccess("foo@example.com/bar")
	assertAcc("foo@example.com/Access", e, err)

	// A directory should return the access file it contains
	e, err = d.WhichAccess("foo@example.com/")
	assertAcc("foo@example.com/Access", e, err)

	// A non-existent path should return the nearest access file along the path
	e, err = d.WhichAccess("foo@example.com/baz/qux/quux")
	assertAcc("foo@example.com/Access", e, err)

	// A link should return ErrFollowLink

	// An access file should return itself

	// When no access file exists it should return nil
}
