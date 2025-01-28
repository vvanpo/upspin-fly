package dirserver

import (
	"context"

	"upspin.io/access"
	"upspin.io/errors"
	"upspin.io/serverutil"
	"upspin.io/upspin"
)

// Glob implements upspin.DirServer.
func (d *dialed) Glob(pattern string) ([]*upspin.DirEntry, error) {
	ctx, op := d.setCtx("Glob")
	d.log = d.log.With("pattern", pattern)

	lookup := func(name upspin.PathName) (*upspin.DirEntry, error) {
		return d.lookupContext(ctx, op, name)
	}
	ls := func(name upspin.PathName) ([]*upspin.DirEntry, error) {
		return d.list(ctx, op, name)
	}

	// TODO multiple consecutive lookups should be in a transaction, to prevent
	// race conditions with Delete(). But access checks involve ipc (multiple
	// for remote group checks), ~~so perhaps the best solution is to perform the
	// glob as though read+list rights are granted, queue the various access
	// checks, and once they've completed elide what's needed from the result.~~
	// Or, change the State interface to return and accept opaque entry
	// identifiers on lookups which map to the append-only log, which foregoes
	// the need for a transaction. This requires a tree of entries navigable
	// from the root for every sequence update, now I understand why the
	// reference implementation uses a Merkle tree.
	es, err := serverutil.Glob(pattern, lookup, ls)
	if err != nil && err != upspin.ErrFollowLink {
		// list() returns errors decorated with op, but serverutil.Glob()
		// creates some of its own.
		return nil, errors.E(op, err)
	}

	for _, e := range es {
		if e.IsRegular() && !e.IsIncomplete() {
			// TODO add blocks to entry
		}
	}

	return es, err
}

// list implements the functionality required by serverutil.ListFunc. Returns
// incomplete entries, but only marks them as such if the requester does not
// have read access to them.
//
// Returns errors.Private, .Permission, .NotExist, upspin.ErrFollowLink, or an
// internal error.
func (d *dialed) list(ctx context.Context, op errors.Op, name upspin.PathName) ([]*upspin.DirEntry, error) {
	// TODO serverutil.Glob() performs repeated lookups with list() (once per
	// metacharacter in the pattern), so successive calls to lookup() include
	// redundant partial lookups of the base path entries that were already
	// retrieved. This could be solved by closing over a map of paths ->
	// EntryIds.
	p, e, a, _, err := d.lookup(ctx, name)
	if err == upspin.ErrFollowLink {
		return []*upspin.DirEntry{e}, err
	} else if errors.Is(errors.Invalid, err) || errors.Is(errors.Private, err) {
		return nil, errors.E(op, p.Path(), err)
	} else if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	}

	canList, err := d.can(ctx, a, access.List, p)
	if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	} else if !canList {
		return nil, errors.E(op, errors.Permission)
	}

	if !e.IsDir() {
		// For Glob requests of `name + "/*"` or similar, if name is a regular
		// file then there are no entries matching the pattern.
		return nil, nil
	}

	es, err := d.state.List(ctx, p.Path())
	if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	}

	// Read access applies uniformly for files within a directory.
	canRead, err := d.can(ctx, a, access.Read, p)
	if err != nil {
		return nil, d.internalErr(ctx, op, p.Path(), err)
	}

	if !canRead {
		for _, e := range es {
			// Only regular files are marked as incomplete.
			if e.IsRegular() && !access.IsAccessControlFile(e.Name) {
				e.MarkIncomplete()
			}
		}
	}

	return nil, nil
}
