package dirserver

import "upspin.io/upspin"

// Watch implements upspin.DirServer.
func (d *dialed) Watch(name upspin.PathName, sequence int64, done <-chan struct{}) (<-chan upspin.Event, error) {
	return nil, upspin.ErrNotSupported
}
