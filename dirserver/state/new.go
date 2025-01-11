package state

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schema string

func New(db *sql.DB) (State, error) {
	return State{db}, nil
}

func (s State) create(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, stmt := range strings.Split(schema, ";\n") {
		_, err = tx.ExecContext(ctx, stmt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
