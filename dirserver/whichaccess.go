package dirserver

import (
	"context"
	"log/slog"

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
	- if the user has no read rights, return the entry without blocks or packdata

TODO snapshot trees; always return nil
TODO if a request with a non-existent path contains an ancestor element that is a file instead of a directory, what should be returned?
	- the reference implementation just pretends it's a folder; e.g. WhichAccess("user@example.com/Access/Access") (which is never a possible path) will return "user@example.com/Access" if it exists.
*/

// WhichAccess implements upspin.DirServer.
func (s *server) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	const op errors.Op = "dir.WhichAccess" // TODO put op, requester, path, etc. in request context
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
	ae, err := s.accessFor(ctx, e)
	if err != nil {
		return nil, s.internalErr(ctx, op, name, err)
	}

	var a *access.Access
	if ae != nil {
		a, err = s.Cache.GetAccess(ctx, ae)
		if err != nil {
			// If the access file is malformed or the store server serving it
			// is down or can't be reached, we don't want the directory server
			// to be unusable, so we pretend the access file isn't there and
			// fall back on the default owner-only rights.
			s.Logger.ErrorContext(
				ctx,
				"Access file cannot be retrieved or parsed",
				slog.String("op", string(op)),
				slog.String("name", string(name)),
				slog.Any("err", err),
			)
		}
	}

	if a == nil {
		// TODO remove and update access.Can() to allow nil receiver as a shortcut for an owner check
		a, _ = access.New(p.First(0).Path())
	}

	getGroup := func(n upspin.PathName) ([]byte, error) {
		return s.Cache.GetGroup(ctx, n)
	}
	if granted, err := a.Can(requester, access.AnyRight, p.Path(), getGroup); err != nil {
		s.Logger.ErrorContext(
			ctx,
			"access check failed",
			slog.String("right", access.AnyRight.String()),
			slog.String("op", string(op)),
			slog.String("name", string(name)),
			slog.Any("err", err),
		)
	} else if !granted {
		return nil, errors.E(op, name, errors.Private)
	}

	if e.IsLink() {
		return e, upspin.ErrFollowLink
	}

	if ae == nil {
		return nil, nil
	}

	if canRead, err := a.Can(requester, access.Read, ae.Name, getGroup); err != nil {
		s.Logger.ErrorContext(
			ctx,
			"access check failed",
			slog.String("right", access.Read.String()),
			slog.String("op", string(op)),
			slog.String("name", string(name)),
			slog.Any("err", err),
		)
	} else if !canRead {
		ae.MarkIncomplete()
	}

	return ae, nil
}

// Returns the access file entry defining access rules for the passed entry.
// Does not follow links.
func (s *server) accessFor(ctx context.Context, e *upspin.DirEntry) (*upspin.DirEntry, error) {
	p, _ := path.Parse(e.Name)
	if e.Attr != upspin.AttrDirectory {
		p = p.Drop(1)
	}

	var ae *upspin.DirEntry
	var err error
	for i := p.NElem(); i >= 0; i++ {
		dir := p.Path()
		if !p.IsRoot() {
			dir += "/"
		}

		ae, err = s.State.Lookup(ctx, dir+access.AccessFile)
		if err != nil || ae != nil {
			break
		}

		p = p.Drop(1)
	}

	return ae, nil
}
