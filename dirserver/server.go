package dirserver

import (
	"context"
	"log/slog"

	"upspin.io/errors"
	"upspin.io/path"
	"upspin.io/upspin"
)

// State provides a persistence interface for all data managed by the directory
// server.
//
// Returned errors should generally be regarded as internal faults.
type State interface {

	// LookupAll retrieves the entries for all elements in a path. If a link is
	// found in the path it is the last element returned, regardless of whether
	// this completes the requested path. If an entry does not exist, the
	// elements up to and including its nearest existing parent are returned.
	// If a regular file entry is returned, it is marked incomplete and
	// contains no blocks.
	LookupAll(context.Context, path.Parsed) ([]*upspin.DirEntry, error)

	// Lookup retrieves the entry at the requested path, if it exists. It does
	// not attempt to evaluate links along the path.
	Lookup(context.Context, path.Parsed) (*upspin.DirEntry, error)

	// GetBlocks returns the blocks for an entry at the given path. Performs no
	// validation; the entry must exist and be a regular file.

	// Put persists a put operation. Performs no validation; all intermediate
	// elements must exist and be directories or it will result in state
	// corruption.
	Put(context.Context, *upspin.DirEntry) error

	// Delete persists a delete operation for the entry at a given path.
	// Performs no validation; the entry must exist (and must not be a
	// directory with children) or will result in an inconsistent state.
	Delete(context.Context, path.Parsed) error
}

// Cache provides an interface for transparent caching of all data depended on
// by the directory server that it does not manage, i.e. access and group file
// contents.
type Cache interface {
}

type server struct {
	State
	*slog.Logger
}

// Logs and formats an internal error to pass to the user, eliding details.
func (s *server) internalErr(ctx context.Context, msg string, op errors.Op, name upspin.PathName, err error) error {
	s.Logger.ErrorContext(
		ctx,
		msg,
		slog.String("op", string(op)),
		slog.String("name", string(name)),
		slog.Any("err", err),
	)

	return errors.E(op, name, errors.Internal, "reference: TODO") // Figure out tracing/correlation ids
}
