package sqlite

import (
	"database/sql"
	"strings"

	"upspin.io/path"
)

// Updates a path in the cache.
func cachePut(tx *sql.Tx, p path.Parsed, op int64) error {
	r := tx.QueryRow(
		`SELECT sequence
		FROM cache_entry
		WHERE name = ?`,
		p.First(0).Path(),
	)
	var seq int64
	if err := r.Scan(&seq); err != nil {
		if err == sql.ErrNoRows {
			seq = 0
		} else {
			return err
		}
	}
	seq++

	// Upsert the final entry
	_, err := tx.Exec(
		`INSERT INTO cache_entry
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			op = excluded.op,
			sequence = excluded.sequence`,
		p.Path(),
		op,
		seq,
	)
	if err != nil {
		return err
	}
	if p.NElem() == 0 {
		return nil
	}

	// Update sequence of all intermediary directories
	dirs := make([]any, p.NElem())
	for i := 0; i < p.NElem(); i++ {
		dirs[i] = p.First(i).String()
	}
	params := strings.Repeat("?,", len(dirs))
	params = params[:len(params)-1]
	bind := append([]any{seq}, dirs...)
	_, err = tx.Exec(
		`UPDATE cache_entry
		SET sequence = ?
		WHERE name IN (`+params+`)`,
		bind...,
	)

	return err
}

// Deletes a path from the cache.
func cacheDelete(tx *sql.Tx, p path.Parsed) error {
	return nil
}

// Sets the sequence of all elements in the path to an incremented sequence and
// returns it.
// e.g. if
// path:	user@example.com/foo/bar/baz
// sequences:             34  18   7   6
// the resulting sequences for all elements will be 35.
func cacheUpdateSeq(tx *sql.Tx, p path.Parsed) (int64, error) {
	return -1, nil
}
