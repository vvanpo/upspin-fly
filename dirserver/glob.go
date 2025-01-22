package dirserver

import (
	"context"

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
	// for remote group checks), so perhaps the best solution is to perform the
	// glob as though read+list rights are granted, queue the various access
	// checks, and once they've completed elide what's needed from the result.
	return serverutil.Glob(pattern, lookup, ls)
}

// list implements the functionality required by serverutil.ListFunc.
func (d *dialed) list(ctx context.Context, op errors.Op, name upspin.PathName) ([]*upspin.DirEntry, error) {
	return nil, nil
}
