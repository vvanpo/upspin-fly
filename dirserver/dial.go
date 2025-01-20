package dirserver

import "upspin.io/upspin"

// Dial implements upspin.Dialer.
func (s *server) Dial(rc upspin.Config, e upspin.Endpoint) (upspin.Service, error) {
	d := &dialed{
		server:    s,
		requester: rc.UserName(),
	}

	return d, nil
}

// Endpoint implements upspin.Service.
func (d *dialed) Endpoint() upspin.Endpoint {
	return d.server.cfg.DirEndpoint()
}

// Close implements upspin.Service.
func (d *dialed) Close() {
}
