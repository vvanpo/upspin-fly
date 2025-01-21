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
	listDir := func(name upspin.PathName) ([]*upspin.DirEntry, error) {
		return d.listDir(ctx, op, name)
	}

	return serverutil.Glob(pattern, lookup, listDir)
}

func (d *dialed) listDir(ctx context.Context, op errors.Op, name upspin.PathName) ([]*upspin.DirEntry, error) {
	return nil, nil
}
