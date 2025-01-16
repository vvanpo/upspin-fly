package sqlite

import (
	"database/sql"
	"strings"

	"upspin.io/path"
)

// Updates a path in the cache.
func cachePut(tx *sql.Tx, p path.Parsed, op int64) error {
	var seq int64 = 1

	if !p.IsRoot() {
		var err error
		seq, err = cacheUpdateSeq(tx, p.Drop(1))
		if err != nil {
			return err
		}
	}

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

	return err
}

// Deletes a path from the cache.
func cacheDelete(tx *sql.Tx, p path.Parsed) error {
	_, err := tx.Exec(
		`DELETE FROM cache_entry
		WHERE name = ?`,
		p.Path(),
	)
	if err != nil {
		return err
	}

	_, err = cacheUpdateSeq(tx, p.Drop(1))
	return err
}

// Sets the sequence of all elements in the path to an incremented sequence and
// returns it. All elements including the root directory must exist in the
// cache.
// e.g. if
// path:	user@example.com/foo/bar/baz
// sequences:             34  18   7   6
// the resulting sequences for all elements will be 35.
func cacheUpdateSeq(tx *sql.Tx, p path.Parsed) (int64, error) {
	r := tx.QueryRow(
		`SELECT sequence
		FROM cache_entry
		WHERE name = ?`,
		p.First(0).Path(),
	)
	var seq int64
	if err := r.Scan(&seq); err != nil {
		return -1, err
	}
	seq++

	els := make([]any, p.NElem()+1)
	for i := 0; i < p.NElem()+1; i++ {
		els[i] = p.First(i).String()
	}
	params := strings.Repeat("?,", len(els))
	params = params[:len(params)-1]
	bind := append([]any{seq}, els...)
	_, err := tx.Exec(
		`UPDATE cache_entry
		SET sequence = ?
		WHERE name IN (`+params+`)`,
		bind...,
	)

	return seq, err
}
