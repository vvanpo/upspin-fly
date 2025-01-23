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
- contains any link elements (including the final element), return the link
  closest to the root and upspin.ErrFollowLink if the user has any access right
  on the link, else errors.Private
- does not grant any access rights to the requester, return errors.Private
- does not exist, return errors.NotExist
- exists and...
  - grants the requester read rights or is a special access or group file,
    return the entry
  - grants any right (but not access.Read), return the entry without blocks or
    packing data and marked as incomplete

TODO snapshot access control
*/

// Lookup implements upspin.DirServer.
func (d *dialed) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	ctx, op := d.setCtx("Lookup")
	d.log = d.log.With("pathname", name)
	return d.lookupContext(ctx, op, name)
}

func (d *dialed) lookupContext(ctx context.Context, op errors.Op, name upspin.PathName) (*upspin.DirEntry, error) {
	p, e, a, _, err := d.lookup(ctx, name)
	if err == upspin.ErrFollowLink {
		return e, err
	} else if errors.Is(errors.Invalid, err) || errors.Is(errors.Private, err) {
		return nil, errors.E(op, p.Path(), err)
	} else if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	}

	canRead, err := d.can(ctx, a, access.Read, p)
	if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	} else if !canRead && !access.IsAccessControlFile(e.Name) {
		e.MarkIncomplete()
	} else {
		// TODO populate blocks+packdata
	}

	return e, nil
}

// lookup is a convenience method that returns the parsed input path, the
// directory entry if it's found, the access file controlling access to the
// entry, the corresponding directory entry for the access file, and a possible
// error. If the path does not exist, nil is returned. The returned entry for
// the input path is incomplete (i.e. without blocks or packing) but not marked
// as such. The returned access file entry is always complete if present.
//
// If the requested pathname is invalid, errors.Invalid is returned.
// If the requesting user has no access rights on the pathname, errors.Private
// is returned.
// If a link is found anywhere along the path, upspin.ErrFollowLink is
// returned.
// All other returned errors are unsanitized internal errors.
//
// TODO return only sanitized errors.
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
			// TODO distinguish between error in access file fetching (warning)
			// and parsing (error)
			d.log.ErrorContext(
				ctx,
				"access file retrieval failed",
				"err", err,
			)
			// If the access file is malformed or the store server serving it
			// can't be reached, we don't want the directory server to be
			// unusable, so we pretend the access file isn't there and fall
			// back on the default owner-only rights.
		}
	}

	if granted, err := d.can(ctx, a, access.AnyRight, p); err != nil {
		return p, nil, nil, nil, err
	} else if !granted {
		return p, nil, nil, nil, errors.E(errors.Private)
	}

	if e.IsLink() {
		return p, e, nil, nil, upspin.ErrFollowLink
	}

	return p, e, a, ae, nil
}
