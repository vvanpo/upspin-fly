package dirserver

import (
	"upspin.io/upspin"
)

// Dial implements upspin.Dialer.
func (s *server) Dial(rc upspin.Config, e upspin.Endpoint) (upspin.Service, error) {
	requester := rc.UserName()
	d := &dialed{
		server:    s,
		log:       s.log.With("requester", requester),
		requester: requester,
	}

	return d, nil
}

// Endpoint implements upspin.Service.
func (d *dialed) Endpoint() upspin.Endpoint {
	return d.server.cfg.DirEndpoint()
}

// Close implements upspin.Service.
func (d *dialed) Close() {
	d.log.Info("closed")
}
