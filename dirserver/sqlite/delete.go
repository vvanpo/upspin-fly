package sqlite

import (
	"context"
	"fmt"

	"upspin.io/path"
)

// Delete implements dirserver.State.
func (s State) Delete(ctx context.Context, p path.Parsed) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction for Delete: %w", err)
	}

	if _, err := s.appendOp(tx, p, -1); err != nil {
		tx.Rollback()
		return fmt.Errorf("persist delete to log: %w", err)
	}

	if err := cacheDelete(tx, p); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete cache entry: %w", err)
	}

	return tx.Commit()
}
