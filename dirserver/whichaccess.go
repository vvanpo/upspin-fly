package dirserver

import (
	"context"

	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

/*
If the WhichAccess path...
- contains any link elements (including the final element), return the link closest to the root and upspin.ErrFollowLink if the user has any access right on the link, else errors.Private
- look for the nearest directory entry on the path (including the final element) that contains an entry named Access, regardless of whether the path or some of its ancestors exist, and..
  - if an Access file is found, and the user has any access right described within, return the entry for the Access file, else errors.Private
  - if no Access file is found and the user is the owner of the tree, return nil, else errors.Private
*/

func (s *server) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	const op errors.Op = "dirserver.WhichAccess"
	ctx := context.TODO()

	p, err := path.Parse(name)
	if err != nil {
		return nil, errors.E(op, name, err)
	}

	_, err = s.State.LookupAll(ctx, p)
	if err != nil {
		return nil, s.internalErr(ctx, "LookupAll", op, name, err)
	}

	return nil, nil
}
