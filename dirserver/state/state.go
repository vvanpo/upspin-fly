// Defines persistence-related interfaces needed by dirserver.
package state

import (
	"context"

	"upspin.io/access"
	"upspin.io/path"
	"upspin.io/upspin"
)

// EntryId is an opaque identifier returned and accepted by State
// implementations that represent immutable entry records in the state. Only
// EntryIds returned by a given State instance should ever be passed back.
type EntryId int64

// State provides a persistence interface for all data managed by the directory
// server.
//
// Returned errors should generally be regarded as internal faults.
type State interface {

	// TODO LookupAll -> LookupElem and return only a slice of found path
	// elements + the final element's attributes and an opaque id for the
	// immutable record. Lookup should be split to LookupPath and LookupIds.
	// List should only take an EntryId.

	// LookupElem finds the nearest element in the passed path that matches an
	// entry in the tree. If the returned path is equal to the passed path, the
	// path exists. The returned upspin.Attribute and EntryId match the final
	// element in the returned path. Attribute is only ever one of AttrNone,
	// AttrLink, or AttrDirectory. A negative EntryId indicates an error or
	// that the root does not exist on this server for the requested path.
	LookupElem(context.Context, path.Parsed) (EntryId, path.Parsed, upspin.Attribute, error)

	// LookupAll retrieves the entries for all elements in a path. If a link is
	// found in the path it is the last element returned, regardless of whether
	// this completes the requested path. If an entry does not exist, the
	// elements up to and including its nearest existing parent are returned.
	// If a regular file entry is returned, it contains packdata without
	// blocks, but is not marked incomplete.
	LookupAll(context.Context, path.Parsed) ([]*upspin.DirEntry, error)

	// Lookup retrieves the entry at the requested path, if it exists. It does
	// not attempt to evaluate links along the path. The path should be clean
	// or the lookup will return nil. If the entry is a file it is returned
	// complete.
	Lookup(context.Context, upspin.PathName) (*upspin.DirEntry, error)

	// List retrieves all entries contained in the directory represented by the
	// entry identifier. If the identifier does not represent a directory, the
	// lookup will return no entries. File entries are returned with packdata
	// but no blocks, yet are not marked incomplete.
	List(context.Context, EntryId) ([]*upspin.DirEntry, error)

	// GetBlocks returns the blocks for an entry at the given path. Performs no
	// validation; the entry must exist and be a regular file.

	// Put persists a put operation. Performs no validation; all intermediate
	// elements must exist and be directories or it will result in state
	// corruption.
	Put(context.Context, *upspin.DirEntry) error

	// Delete persists a delete operation for the entry at a given path.
	// Performs no validation; the entry must exist (and must not be a
	// directory with children) or will result in an inconsistent state.
	Delete(context.Context, path.Parsed) error
}

// Cache provides an interface for transparent caching of all data depended on
// by the directory server that is stored elsewhere, i.e. access and group file
// contents.
type Cache interface {

	// GetAccess retrieves and caches a parsed access file.
	GetAccess(context.Context, *upspin.DirEntry) (*access.Access, error)

	// GetGroup retrieves a local or remote group file. Must be passed to every
	// invocation of access.Can().
	//
	// Currently this method does not live up to the interface's promise;
	// upspin.io/access uses its own global group cache, which this and
	// RemoveGroup will have to be implemented to manipulate.
	GetGroup(context.Context, upspin.PathName) ([]byte, error)

	RemoveGroup(context.Context, upspin.PathName) error
}
