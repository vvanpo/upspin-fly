-- The tables prefixed with `log_` are append-only, and its records are
-- immutable. They serve as the source-of-truth for the state of each user tree
-- managed by this server.

CREATE TABLE log_root (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	username TEXT UNIQUE NOT NULL
);

CREATE TABLE log_put (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	writer TEXT NOT NULL,
	-- If true, the below fields are not present
	dir BOOLEAN DEFAULT FALSE NOT NULL,
	-- If not null, the below fields are not present
	link TEXT,
	packing INTEGER,
	packdata BLOB
);

CREATE TABLE log_operation (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	timestamp INTEGER DEFAULT (unixepoch()) NOT NULL,
	root REFERENCES log_root NOT NULL,
	-- Path under the root directory, without the username or leading /
	path TEXT NOT NULL,
	-- If null, implies this operation is a deletion
	put REFERENCES log_put UNIQUE
);

CREATE TABLE log_block (
	put REFERENCES log_put NOT NULL,
	endpoint TEXT NOT NULL,
	reference TEXT NOT NULL,
	offset INTEGER NOT NULL,
	size INTEGER NOT NULL,
	packdata BLOB,
	PRIMARY KEY(put, reference)
);

-- Represents the current state of tree as projected from the log history. Can
-- be computed by replaying the log, but is kept in sync with every put or
-- delete operation to serve as a cache of the current sequence.
CREATE TABLE proj_entry (
	name TEXT PRIMARY KEY NOT NULL,
	-- This must reference an op with a non-null `put` column
	op REFERENCES log_operation UNIQUE NOT NULL,
	sequence INTEGER NOT NULL,
	-- The parent directory. Only the root path of a tree references itself.
	parent REFERENCES log_put NOT NULL
);
