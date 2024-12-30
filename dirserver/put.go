package dirserver

import "upspin.io/upspin"

/*
If the Put entry...
- parent path contains a link, return ErrFollowLink if the user has any access right for the first link in the path
- replaces an existing entry:
  - the existing entry can't be a directory
  - the new entry can't be a directory
  - the user must have the Write permission at the path
  - the entry sequence number must match the existing entry, or be SeqIgnore
- is a new entry:
  - the user must have the Create permission at the path
  - the sequence number must be SeqNotExist or SeqIgnore
- path last element is Access:
  - must be a regular file
- path is <user>/Group:
  - must be a directory
- path is within the subtree rooted at <user>/Group:
  - cannot be a link
  - path elements cannot resemble a username
- is a special file (Access or /Group/...):
  - the user must be the owner
  - must use signed-but-unencrypted packing
*/

func (s *server) Put(entry *upspin.DirEntry) (*upspin.DirEntry, error) {

	return &upspin.DirEntry{
		Attr:     upspin.AttrIncomplete,
		Sequence: 1,
	}, nil
}
