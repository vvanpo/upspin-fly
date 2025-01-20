package dirserver

import "upspin.io/upspin"

// Glob implements upspin.DirServer.
func (d *dialed) Glob(pattern string) ([]*upspin.DirEntry, error) {
	return nil, upspin.ErrNotSupported
}
