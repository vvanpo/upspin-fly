package sqlite

import (
	"context"

	"upspin.io/upspin"
)

func (s State) List(context.Context, upspin.PathName) ([]*upspin.DirEntry, error) {
	return nil, nil
}
