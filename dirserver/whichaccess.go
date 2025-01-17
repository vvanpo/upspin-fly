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
  - if an Access file is found, and the user has any access right described within, return the entry for the Access file, else errors.Private
	- if the user has no read rights, return the entry without blocks or packdata
  - if no Access file is found and the user is the owner of the tree, return nil, else errors.Private

TODO: snapshot users; always return nil
*/

// WhichAccess implements upspin.DirServer.
func (s *server) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	const op errors.Op = "dirserver.WhichAccess"
	ctx := context.TODO()

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	es, err := s.State.LookupAll(ctx, p)
	if err != nil {
		return nil, s.internalErr(ctx, op, name, err)
	}

	e := es[len(es)-1]
	acc, err := s.accessFor(ctx, e)
	if err != nil {
		return nil, s.internalErr(ctx, op, name, err)
	}

	if e.Attr == upspin.AttrLink {
		return s.followLink(e)
	}
	if acc == nil {
		// TODO check if owner
		// return nil, errors.E(op, name, errors.Private)
	}

	// TODO elide blocks+packdata from entry if user does not have access.Read
	return acc, nil
}

// Returns the access file controlling access to the path represented by the
// entry list. Does not follow links.
func (s *server) accessFor(ctx context.Context, e *upspin.DirEntry) (*upspin.DirEntry, error) {
	p, _ := path.Parse(e.Name)
	if e.Attr != upspin.AttrDirectory {
		p = p.Drop(1)
	}

	var ae *upspin.DirEntry
	for i := p.NElem(); i >= 0; i++ {
		ap := p.Path() + "/" + access.AccessFile
		ae, err := s.State.Lookup(ctx, ap)
		if err != nil {
			return nil, err
		}
		if ae != nil {
			break
		}

		if p.IsRoot() {
			break
		}
		p = p.Drop(1)
	}

	return ae, nil
}

func (s *server) followLink(l *upspin.DirEntry) (*upspin.DirEntry, error) {
	// TODO check Any access
	// return nil, errors.E(op, name, errors.Private)

	return l, upspin.ErrFollowLink
}
