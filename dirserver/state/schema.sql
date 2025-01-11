CREATE TABLE root (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	username TEXT UNIQUE NOT NULL
);

CREATE TABLE log_operation (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	timestamp INTEGER DEFAULT CURRENT_TIMESTAMP NOT NULL,
	root REFERENCES root NOT NULL,
	-- path under the root directory, without the username or leading /
	path TEXT NOT NULL
);

CREATE TABLE log_put (
	operation REFERENCES log_operation PRIMARY KEY NOT NULL,
	writer TEXT,
	-- if not null, the other fields are not present
	link TEXT,
	dir BOOLEAN,
	packing INTEGER,
	packdata BLOB
);

CREATE TABLE log_delete (
	operation REFERENCES log_operation PRIMARY KEY NOT NULL
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
