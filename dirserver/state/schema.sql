CREATE TABLE root (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	username TEXT UNIQUE NOT NULL
);

CREATE TABLE log_put (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	writer TEXT NOT NULL,
	-- if not null, the below fields are not present
	link TEXT,
	-- if true, link and the below fields are not present
	dir BOOLEAN,
	packing INTEGER,
	packdata BLOB
);

CREATE TABLE log_operation (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	timestamp INTEGER DEFAULT CURRENT_TIMESTAMP NOT NULL,
	root REFERENCES root NOT NULL,
	-- path under the root directory, without the username or leading /
	path TEXT NOT NULL,
	-- if null, implies this operation is a deletion
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
