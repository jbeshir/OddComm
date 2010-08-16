package core

// Trie represents an iterable radix tree.
// It is designed for storing items with minimal overhead in small cases,
// and acceptable overhead in large ones, iterable without invalidating
// iterators in any operation, and searchable by key prefix.
// An "empty" trie in this package is just a nil pointer of this type.
// This container is not threadsafe; only one operation should be performed
// upon it at once.
type Trie struct {
	down *Trie        // First child of this node.
	next *Trie        // Next sibling of this node- or the next one back
		          // up if it has none left.
	nodekey string    // Value of this node.
	key string        // Full key of this node, if it has a key.
        value interface{} // If set, this node contains a value.
	lastSib bool      // Whether this node is the last of its siblings.
}

// Iterator safety: Thanks to garbage collection, we get 90% of this for free.
// If we have a pointer to a node, then all its pointers will still be valid.
// Eventually, we will return to a part of the trie that still exists, and,
// if no one else is pointing to the nodes we iterated, they will be deleted.
// The only gotcha is that we cannot change the parent node of a set of nodes,
// as this would cause us to miss our 'root' node while iterating a subtrie.

// Value returns this trie node's key and value, or nil if it is not a value
// node.
func (t *Trie) Value() (string, interface{}) {
	return t.key, t.value
} 

// Next gets the next value node, not including this node itself if it has
// a value. If end is non-nil, it indicates a node to not go past, used to
// iterate only within a subtrie.
func (t* Trie) Next(end *Trie) (next *Trie) {

	next = nextNode(end, t)
	for next != end && next != nil && next.key == "" {
		next = nextNode(end, next)
	}
	
	// If we found the end node, return nil, not it.
	if next == end {
		next = nil
	}

	return
}

// Iterate to the next node, including non-value nodes. May return nil.
func nextNode(end, current *Trie) *Trie {
	if current.down != nil {
		current = current.down
	} else {
		for cont := true; cont && current != nil; {

			// If this is the end, we do not want to go to its
			// next node, so return nil.
			if current == end {
				return nil
			}

			// If this was a last sibling, we're going back up to
			// its parent; continue onto ITS next node, not
			// recursing into children.
			cont = false
			if current.lastSib {
				cont = true
			}

			current = current.next
		}
	}
	return current
}

// TrieAdd adds the given key and value to the trie.
// If it already exists, it is overwritten with the new value.
// Returns the previous value of the given key, or nil if it had none.
func TrieAdd(first **Trie, key string, value interface{}) (old interface{}) {
	remaining := key
	n := *first
	var parent *Trie
	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
				n.nodekey[i] == remaining[i]; i++ {}

		// If this key has nothing in common with ours, next.
		if i == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.lastSib {
				n = nil
				break
			}

			// Otherwise, continue to the next node.
			n = n.next
			continue
		}

		// If we matched the whole of the key...
		if i == len(n.nodekey) {

			// Exact match? Make it a data node if it wasn't
			// already, and set our value.
			if i == len(remaining) {
				old = n.value
				n.key = key
				n.value = value
				return
			}

			// Otherwise, set this as our new root, cut the start
			// off remaining, and restart. Note the parent node.
			parent = n
			first = &n.down
			remaining = remaining[i:]
			n = *first
			continue
		}

		// Otherwise, our key matches part of this node's key, but not
		// all, and we need to split this node into two.
		newn := new(Trie)
		newn.down = n.down
		newn.next = n
		newn.nodekey = n.nodekey[i:]
		newn.lastSib = true
		newn.key = n.key
		newn.value = n.value
		n.down = newn
		n.nodekey = n.nodekey[0:i]
		n.key = ""
		n.value = nil

		// If none of our name is left, then the first parent node
		// above matches it. Otherwise, we need to add ourselves as a
		// new child.
		if len(remaining) == i {
			newn = n
		} else {
			newn = new(Trie)
			newn.next = n.down
			newn.nodekey = remaining[i:]
			n.down = newn
		}
		newn.key = key
		newn.value = value
		return
	}

	// If we've gone through the list and not found anything with
	// a common prefix, we need to simply add ourselves.
	if n == nil {
		n := new(Trie)

		// Set the next node, which is the first sibling. If there
		// isn't one, it's the parent, and we're the last sibling.
		n.next = *first
		if n.next == nil {
			n.next = parent
			n.lastSib = true
		}
		n.nodekey = remaining
		n.key = key
		n.value = value
		(*first) = n
		return
	}
	return
}

// TrieDel deletes a key from the trie.
// If it does not exist, nothing happens.
// Returns the value of the removed key, or nil if it did not exist.
func TrieDel(first **Trie, key string) (old interface{}) {
	remaining := key
	root := first
	n := *root
	var prev *Trie
	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
				n.nodekey[i] == remaining[i]; i++ {}

		// If this key has nothing in common with ours, next.
		if i == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.lastSib {
				n = nil
				break
			}

			// Otherwise, continue to the next node.
			prev = n
			n = n.next
			continue
		}

		// If we didn't match the whole of the key, return.
		// No match.
		if i != len(n.nodekey) {
			return
		}

		// Exact match? If it's a value node, delete it.
		// Otherwise, no match exists.
		if i == len(remaining) {
			if n.key != "" {
				old = n.value
				if n.down != nil && n.down.next != n {
					// This node is in use as part
					// of multiple other keys;
					// just remove its value.
					n.key = ""
					n.value = nil
					return
				}

				if n.down != nil {
					// Merge our one child into us.
					// This HAS to happen this way around,
					// or we'd be editing the child's
					// parent (breaks iterators).
					n.nodekey = n.nodekey + n.down.nodekey
					n.key = n.down.key
					n.value = n.down.value
					n.down = n.down.down
					return
				}

				// No children, so just delete references to
				// us.
				if prev != nil {
					prev.next = n.next
					prev.lastSib = n.lastSib
				} else {
					// If we had siblings, just remove us.
					if !n.lastSib {
						*root = n.next
						return
					}

					// Otherwise, recurse up, deleting
					// parents without children or values.
					n.key = ""; n.value = nil
					for ; n != nil; n = n.next {
						// If we have a value, stop.
						if n.key != "" {
							return
						}

						// If we have reached the top
						// of the tree, stop, after
						// deleting ourselves if we're
						// the sole item left.
						if n.next == nil {
							if *first == n {
								*first = nil
							}
							return
						}

						// If we are not the sole
						// sibling, stop recursing.
						if n.next.down != n {
							return
						}

						// Otherwise, delete ourselves.
						n = n.next
						n.down = nil
					}
				}
				return
			}
			return
		}

		// Otherwise, set this as our new root,
		// cut the start off remaining, and restart.
		root = &n.down
		remaining = remaining[i:]
		prev = nil
		n = *root
	}

	// Failed to find anything, return.
	return
}


// TrieGet gets a key's value from the trie.
// If it does not exist, returns nil.
func TrieGet(first **Trie, key string) (value interface{}) {
	remaining := key
	n := *first
	for n != nil {

		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
				n.nodekey[i] == remaining[i]; i++ {}

		// If this key has nothing in common with ours, next.
		if i == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.lastSib {
				n = nil
				break
			}

			// Otherwise, continue to the next node.
			n = n.next
			continue
		}

		// If we didn't match the whole of the key, return.
		// No match.
		if i != len(n.nodekey) {
			return
		}

		// Exact match? If it's a value node, return it.
		// Otherwise, no match exists.
		if i == len(remaining) {
			if n.key != "" {
				value = n.value
			}
			return
		}

		// Otherwise, set this as our new root,
		// cut the start off remaining, and restart.
		first = &n.down
		remaining = remaining[i:]
		n = *first
	}

	// Failed to find anything, return.
	return
}

// TrieGetSub gets a subtrie from the trie, consisting of a root of all values
// with a given prefix. The value of the chain up to subtrie returned may be
// longer than the prefix given, if no other sub entries exist.
// The only safe operation on a subtrie is iterating it with Next(). The Trie*
// functions require that they be working on keys from the start of the tree.
// Returns nil if there are no entries with the given prefix.
func TrieGetSub(first **Trie, prefix string) (subtree *Trie) {
	remaining := prefix
	n := *first
	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
				n.nodekey[i] == remaining[i]; i++ {}

		// If this key has nothing in common with ours, next.
		if i == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.lastSib {
				n = nil
				break
			}

			// Otherwise, continue to the next node.
			n = n.next
			continue
		}

		// Either is our prefix, or fully contains it?
		// Return it.
		if i == len(remaining) {
			subtree = n
			return
		}

		// Otherwise, if we didn't match the whole of the key,
		// return. No match.
		if i != len(n.nodekey) {
			return
		}

		// Otherwise, set this as our new root,
		// cut the start off remaining, and restart.
		first = &n.down
		remaining = remaining[i:]
		n = *first
	}

	// Failed to find anything, return.
	return
}
