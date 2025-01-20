package dirserver

import "upspin.io/upspin"

// Delete implements upspin.DirServer.
func (d *dialed) Delete(name upspin.PathName) (*upspin.DirEntry, error) {
	return nil, upspin.ErrNotSupported
}
