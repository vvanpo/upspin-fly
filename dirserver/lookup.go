package dirserver

import (
	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

// Lookup implements upspin.DirServer.
func (d *dialed) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	ctx, op := d.setCtx("Lookup", name)

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}
	es, err := d.state.LookupAll(ctx, p)
	if err != nil {
		return nil, d.internalErr(ctx, op, name, err)
	}
	e := es[len(es)-1]
	if len(es) <= p.NElem() {
		if e.IsLink() {
			return e, upspin.ErrFollowLink
		}
		return nil, errors.E(op, name, errors.NotExist)
	}

	return e, nil
}
