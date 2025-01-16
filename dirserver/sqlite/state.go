package sqlite

import (
	"database/sql"
	_ "embed"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"upspin.io/path"
)

//go:embed schema.sql
var schema string

type State struct {
	db *sql.DB
}

// Open accepts a SQLite database file path and initializes it, creating the
// schema if not present.
func Open(p string) (*State, error) {
	db, err := sql.Open("sqlite3", "file:"+p+"?_fk=true")
	if err != nil {
		return nil, err
	}
	s := &State{db}

	//TODO only if not present
	if err := s.create(); err != nil {
		return nil, err
	}
	//TODO prepare statements

	return s, nil
}

// Close closes the database.
func (s State) Close() error {
	return s.db.Close()
}

func (s State) create() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, stmt := range strings.Split(schema, ";\n") {
		_, err = tx.Exec(stmt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// appendOp appends an operation to the log and returns its id. pid is the id
// of the corresponding log_put record, <0 indicates a deletion operation.
func (s State) appendOp(tx *sql.Tx, p path.Parsed, pid int64) (int64, error) {
	var r sql.Result
	var err error
	if pid < 0 {
		r, err = tx.Exec(
			`INSERT INTO log_operation (root, path) VALUES ((SELECT id FROM root WHERE username = ?), ?)`,
			p.User(),
			p.FilePath(),
		)
	} else {
		r, err = tx.Exec(
			`INSERT INTO log_operation (root, path, put) VALUES ((SELECT id FROM root WHERE username = ?), ?, ?)`,
			p.User(),
			p.FilePath(),
			pid,
		)
	}
	if err != nil {
		tx.Rollback()
	}

	i, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}

	return i, err
}
