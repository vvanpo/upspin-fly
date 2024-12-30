package dirserver

import (
	"github.com/vvanpo/upspin-fly/dirserver/state"
	"upspin.io/upspin"
)

type server struct {
	config upspin.Config
	state  *state.State
}

func (s *server) Lookup(name upspin.PathName) (*upspin.DirEntry, error) {
	// parsed, err := path.Parse(name)

	return nil, nil
}
