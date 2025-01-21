package dirserver

import (
	"context"
	"log/slog"

	"upspin.io/access"
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
	// not attempt to evaluate links along the path. The path should be clean
	// or the lookup will return nil.
	Lookup(context.Context, upspin.PathName) (*upspin.DirEntry, error)

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
// by the directory server that is stored elsewhere, i.e. access and group file
// contents.
type Cache interface {

	// GetAccess retrieves and caches a parsed access file.
	GetAccess(context.Context, *upspin.DirEntry) (*access.Access, error)

	// GetGroup retrieves a local or remote group file. Must be passed to every
	// invocation of access.Can().
	//
	// Currently this method does not live up to the interface's promise;
	// upspin.io/access uses its own global group cache, which this and
	// RemoveGroup will have to be implemented to manipulate.
	GetGroup(context.Context, upspin.PathName) ([]byte, error)

	RemoveGroup(context.Context, upspin.PathName) error
}

// Implements an upspin.Dialer that returns an upspin.DirServer.
type server struct {
	state State
	cache Cache
	log   *slog.Logger

	// The upspin user the server is running as; used to retrieve access and
	// group file contents.
	cfg upspin.Config
}

// Implements an upspin.DirServer serving a user that must be authenticated.
type dialed struct {
	*server
	log       *slog.Logger
	requester upspin.UserName
}

func (d *dialed) setCtx(op string) (context.Context, errors.Op) {
	op = "dir." + op
	d.log = d.log.With("operation", op)

	return context.TODO(), errors.Op(op)
}

// Logs and formats an internal error to pass to the user, eliding details.
func (d *dialed) internalErr(ctx context.Context, op errors.Op, name upspin.PathName, err error) error {
	d.log.ErrorContext(
		ctx,
		"internal error returned to user",
		"err", err,
	)

	return errors.E(op, name, errors.Internal, "reference: TODO") // Figure out tracing/correlation ids
}
