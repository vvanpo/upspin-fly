package dirserver

import (
	"context"

	"upspin.io/access"
	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

/*
If the WhichAccess path...
- contains any link elements (including the final element), return the link closest to the root and upspin.ErrFollowLink if the user has any access right on the link, else errors.Private
- look for the nearest directory entry on the path (including the final element) that contains an entry named Access, regardless of whether the path or some of its ancestors exist, and..
  - if no Access file is found and the user is the owner of the tree, return nil, else errors.Private
  - if an Access file is found, and the user has any access right described within, return the entry for the Access file, else errors.Private

TODO snapshot trees; always return nil
TODO if a request with a non-existent path contains an ancestor element that is a file instead of a directory, what should be returned?
  - the reference implementation just pretends it's a folder; e.g. WhichAccess("user@example.com/Access/Access") (which is never a possible path) will return "user@example.com/Access" if it exists.
*/

// WhichAccess implements upspin.DirServer.
func (d *dialed) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	ctx, op := d.setCtx("WhichAccess", name)

	_, e, _, ae, err := d.lookup(ctx, name)
	if err == upspin.ErrFollowLink {
		return e, err
	} else if errors.Is(errors.Invalid, err) || errors.Is(errors.Private, err) {
		return nil, errors.E(op, name, err)
	} else if err != nil {
		return nil, d.internalErr(ctx, op, name, err)
	}

	return ae, nil
}

// Returns the access file entry defining access rules for the path.
// Does not follow links.
func (s *server) accessFor(ctx context.Context, p path.Parsed, isDir bool) (*upspin.DirEntry, error) {
	if !isDir {
		p = p.Drop(1)
	}

	var ae *upspin.DirEntry
	var err error
	for i := p.NElem(); i >= 0; i-- {
		dir := p.Path()
		if i != 0 {
			dir += "/"
		}

		ae, err = s.state.Lookup(ctx, dir+access.AccessFile)
		if err != nil || ae != nil {
			break
		}

		p = p.Drop(1)
	}

	return ae, err
}
