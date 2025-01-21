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

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	es, err := d.state.LookupAll(ctx, p)
	if err != nil {
		return nil, d.internalErr(ctx, op, name, err)
	}

	e := es[len(es)-1]
	ae, err := d.accessFor(ctx, p, e.Attr == upspin.AttrDirectory)
	if err != nil {
		return nil, d.internalErr(ctx, op, name, err)
	}

	var a *access.Access
	if ae != nil {
		a, err = d.cache.GetAccess(ctx, ae)
		if err != nil {
			// If the access file is malformed or the store server serving it
			// is down or can't be reached, we don't want the directory server
			// to be unusable, so we pretend the access file isn't there and
			// fall back on the default owner-only rights.
			d.log.ErrorContext(
				ctx,
				"access file cannot be retrieved or parsed",
				"err", err,
			)
		}
	}

	if a == nil {
		// TODO remove and update access.Can() to allow nil receiver as a shortcut for an owner check
		a, _ = access.New(p.First(0).Path())
	}

	getGroup := func(n upspin.PathName) ([]byte, error) {
		return d.cache.GetGroup(ctx, n)
	}
	if granted, err := a.Can(d.requester, access.AnyRight, p.Path(), getGroup); err != nil {
		d.log.ErrorContext(
			ctx,
			"access check failed",
			"right", access.AnyRight.String(),
			"err", err,
		)
	} else if !granted {
		return nil, errors.E(op, name, errors.Private)
	}

	if e.IsLink() {
		return e, upspin.ErrFollowLink
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
	for i := p.NElem(); i >= 0; i++ {
		dir := p.Path()
		if !p.IsRoot() {
			dir += "/"
		}

		ae, err = s.state.Lookup(ctx, dir+access.AccessFile)
		if err != nil || ae != nil {
			break
		}

		p = p.Drop(1)
	}

	return ae, nil
}
