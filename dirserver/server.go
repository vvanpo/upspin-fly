package dirserver

import (
	"context"

	"upspin.io/path"
	"upspin.io/upspin"
)

type State interface {
	// GetAll retrieves the entries for all segments in a path. If a link is
	// found in the path it is the last element returned, regardless of whether
	// this completes the requested path. If an entry does not exist, the
	// segments up to and including its parent are returned.
	GetAll(context.Context, path.Parsed) ([]*upspin.DirEntry, error)

	// Put persists a put operation.
	Put(context.Context, *upspin.DirEntry) error

	// Delete persists a delete operation.
	Delete(context.Context, path.Parsed) error
}

type server struct {
	config upspin.Config
	state  State
}
