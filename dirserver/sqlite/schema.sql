CREATE TABLE root (
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
	root REFERENCES root NOT NULL,
	-- Path under the root directory, without the username or leading /
	path TEXT NOT NULL,
	-- If null, implies this operation is a deletion
	put REFERENCES log_put UNIQUE
);

CREATE TABLE block (
	put REFERENCES log_put NOT NULL,
	endpoint TEXT NOT NULL,
	reference TEXT NOT NULL,
	offset INTEGER NOT NULL,
	size INTEGER NOT NULL,
	packdata BLOB,
	PRIMARY KEY(put, reference)
);

-- Caches the reference to the latest put operation representing the entry, and
-- its sequence number as described in https://pkg.go.dev/upspin.io@v0.1.0/upspin#pkg-constants
CREATE TABLE cache_entry (
	name TEXT PRIMARY KEY NOT NULL,
	-- This must reference an op with a non-null `put` column
	op REFERENCES log_operation UNIQUE NOT NULL,
	sequence INTEGER NOT NULL
);
