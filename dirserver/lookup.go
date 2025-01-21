package dirserver

import (
	"context"

	"upspin.io/access"
	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

/*
If the Lookup path...
- contains any link elements (including the final element), return the link closest to the root and upspin.ErrFollowLink if the user has any access right on the link, else errors.Private
- does not grant any access rights to the requester, return errors.Private
- does not exist, return errors.NotExist
- exists and...
  - grants the requester read rights or is a special access or group file, return the entry
  - grants any right (but not access.Read), return the entry without blocks or packing data and marked as incomplete

TODO snapshot access control
*/

// Lookup implements upspin.DirServer.
func (d *dialed) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	ctx, op := d.setCtx("Lookup", name)

	p, e, _, _, err := d.lookup(ctx, name)
	if err == upspin.ErrFollowLink {
		return e, err
	} else if errors.Is(errors.Invalid, err) || errors.Is(errors.Private, err) {
		return nil, errors.E(op, p.Path(), err)
	} else if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	}

	/*if canRead, err := a.Can(d.requester, access.Read, e.Name, getGroup); err != nil {
		d.log.ErrorContext(
			ctx,
			"access check failed",
			"right", access.Read.String(),
			"err", err,
		)
	} else if !canRead {
		ae.MarkIncomplete()
	}*/

	return e, nil
}

// lookup is a convenience method that returns the parsed input path, the
// directory entry if it's found, the access file controlling access to the
// entry, the corresponding directory entry for the access file, and a possible
// error. If the path does not exist, nil is returned. If the returned entry is
// a file, it is incomplete (i.e. without blocks or packing). The returned
// access file entry is always complete.
//
// If the requested pathname is invalid, errors.Invalid is returned.
// If the requesting user has no access rights on the pathname, errors.Private
// is returned.
// If a link is found anywhere along the path, upspin.ErrFollowLink is
// returned.
// All other returned errors are unsanitized internal errors.
func (d *dialed) lookup(ctx context.Context, name upspin.PathName) (
	p path.Parsed,
	e *upspin.DirEntry,
	a *access.Access,
	ae *upspin.DirEntry,
	err error,
) {
	p, err = path.Parse(name)
	if err != nil {
		return p, nil, nil, nil, err
	}

	es, err := d.state.LookupAll(ctx, p)
	if err != nil {
		return p, nil, nil, nil, err
	}

	// The closest existing entry or the entry itself. Could be a link.
	e = es[len(es)-1]
	ae, err = d.accessFor(ctx, p, e.Attr == upspin.AttrDirectory)
	if err != nil {
		return p, nil, nil, nil, err
	}

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
		return p, nil, nil, nil, errors.E(errors.Private)
	}

	if e.IsLink() {
		return p, e, nil, nil, upspin.ErrFollowLink
	}

	return p, e, a, ae, nil
}
