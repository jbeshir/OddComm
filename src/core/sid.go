package core

import "sync"

import "oddcomm/lib/trie"


// The sidTrie stores existing  server IDs, and the 
var sidTrie trie.Trie
var sidMutex sync.Mutex


// NewSID reserves the given SID (three character UID prefix) if it is free,
// associating the given value with it. It returns success or failure.
// Once a SID is reserved, the acquiring package can safely add users with a
// seven-character UID beginning with it.
func NewSID(sid string, value interface{}) (success bool) {
	sidMutex.Lock()

	if value == nil || sidTrie.Get(sid) != nil {
		success = false
	} else {
		sidTrie.Insert(sid, value)
		success = true
	}

	sidMutex.Unlock()
	return
}

// GetSID returns the value associated with the given SID, or nil if the given
// SID is not recognised.
func GetSID(sid string) interface{} {
	return sidTrie.Get(sid)
}

// ReleaseSID releases the given SID, freeing it to be used again later.
func ReleaseSID(sid string) {
	sidMutex.Lock()
	sidTrie.Remove(sid)
	sidMutex.Unlock()
}

// IterateSID runs the given function for every SID and value.
// This can be used to iterate servers.
func IterateSID(f func(sid string, value interface{})) {
	it := sidTrie.Iterate()
	if it != nil {
		for {
			f(it.Value())

			if !it.Next() {
				break
			}
		}
	}
}
