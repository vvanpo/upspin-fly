package dirserver

import (
	"context"

	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

func (s *server) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	const op errors.Op = "dirserver.Lookup"
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
	if len(es) <= p.NElem() {
		if e.Attr == upspin.AttrLink {
			return e, upspin.ErrFollowLink
		}
		return nil, errors.E(op, name, errors.NotExist)
	}

	return e, nil
}
