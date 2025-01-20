package dirserver

import (
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
*/

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
