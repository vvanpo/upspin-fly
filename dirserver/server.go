// Implements a upspin.DirServer.
// The server behaviour differs from the reference implementation in that
// DirEntry's for directories do not contain blocks or packing, and are never
// marked incomplete.
package dirserver

import (
	"context"
	"log/slog"

	"github.com/vvanpo/upspin-fly/dirserver/state"
	"upspin.io/errors"
	"upspin.io/upspin"
)

// Implements an upspin.Dialer that returns an upspin.DirServer.
type server struct {
	state state.State
	cache state.Cache
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
