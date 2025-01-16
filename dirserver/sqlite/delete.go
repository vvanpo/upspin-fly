package sqlite

import (
	"context"

	"upspin.io/path"
)

// Delete implements dirserver.State.
func (s State) Delete(ctx context.Context, p path.Parsed) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := s.appendOp(tx, p, -1); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
