package dirserver

import (
	"context"
	"log/slog"
	"testing"

	"github.com/vvanpo/upspin-fly/dirserver/sqlite"
	"upspin.io/access"
	"upspin.io/upspin"
)

type cache struct {
	access map[upspin.PathName]string
}

func (c *cache) GetAccess(ctx context.Context, e *upspin.DirEntry) (*access.Access, error) {
	as := c.access[e.Name]
	return access.Parse(e.Name, []byte(as))
}
func (_ *cache) GetGroup(ctx context.Context, n upspin.PathName) ([]byte, error) {
	return nil, nil
}
func (_ *cache) RemoveGroup(ctx context.Context, n upspin.PathName) error {
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

	c := &cache{make(map[upspin.PathName]string)}
	c.access["foo@example.com/Access"] = ""
	s := &server{state: st, cache: c}
	d := &dialed{s, slog.Default(), "foo@example.com"}

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
