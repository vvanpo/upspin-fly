package dirserver

import "upspin.io/upspin"

/*
If the WhichAccess path...
- contains a link element (including the final element), return upspin.ErrFollowLink if the user has any access right on the link
- does not exist and...
	- the parent doesn't exist, return errors.Private
- does not exist and
- exists and the user...
	- has Any access right to the path and...
		-
*/

func (s *server) WhichAccess(name upspin.PathName) (*upspin.DirEntry, error) {
	return nil, nil
}
