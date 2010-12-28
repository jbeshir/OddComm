package trie


// If this file is not called base.go, it is an automatically generated
// variation on it, in which case you should not edit directly nor commit it.

// Comphrension of this code will be massively improved by reading the trie's
// design document.

// Iterator safety: Thanks to garbage collection and atomic writes, we get 90%
// of this for free. If we have a pointer to a node, then all its pointers will
// still be valid. Eventually, we will return to a part of the trie that still
// exists, and, if no one else is pointing to the nodes we iterated, they will
// be deleted. The only gotcha is that we cannot change the parent or
// grandparent node, as this would cause us to miss our 'root' node while
// iterating a subtrie.


// Trie provides the implementation of the radix trie.
// An empty trie uses no more space than a nil pointer.
type Trie struct {
	first *nodeBox   // First node pointer.
}

// nodeBox contains a single trie node. It wraps the node, so it can be changed
// atomically without invalidating pointers.
type nodeBox struct {
	n *node
}

// node stores the contents of a trie node.
type node struct {
	nodekey string      // Key of this node, not including parent nodekeys.
	down    *nodeBox    // First child of this node (nilable).
	next    *nodeBox    // Next sibling of this node (nilable).
	key     string      // Full key of this node, if it has a key.
	value   interface{} // If set, this node contains a value.
}


// Iterator permits iteration of a part of a trie. It keeps a stack of visited
// parent nodes to return to, and will iterate only until it reaches the end
// of the trie or returns upwards to the same level it was called on.
// A single iterator may not be used concurrently, but separate iterators may.
type Iterator struct {
	parents []*nodeBox
	it      *nodeBox
}

// Next moves the iterator on to its next value node.
// Returns whether a next value node existed, or the iterator is now nil.
func (i *Iterator) Next() bool {
	i.nextNode()
	for i.it != nil && i.it.n.key == "" {
		i.nextNode()
	}

	if i.it != nil {
		return true
	}
	return false
}

// Iterate to the next node, including non-value nodes.
func (i *Iterator) nextNode() {
	n := i.it.n
	if n.down != nil {
		i.parents = append(i.parents, i.it)
		i.it = n.down
	} else {
		for i.it != nil {
			if n.next != nil {
				i.it = n.next
				break
			}

			if len(i.parents) > 1 {
				i.it = i.parents[len(i.parents)-1]
				i.parents = i.parents[:len(i.parents)-1]
				n = i.it.n
			} else {
				i.it = nil
			}
		}
	}
}

// Value returns the current value of the node pointed to by this iterator.
func (i *Iterator) Value() (name string, value interface{}) {
	n := i.it.n
	return n.key, n.value
}


// Get gets a key's value from the trie.
// If it does not exist, returns nil.
func (t *Trie) Get(key string) (value interface{}) {
	remaining := key
	i := t.first
	for i != nil {
		n := i.n

		// Find out how much of the node key matches ours.
		var c int
		for c = 0; c < len(n.nodekey) && c < len(remaining) &&
			n.nodekey[c] == remaining[c]; c++ {
		}

		// If this key has nothing in common with ours, next.
		if c == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.next == nil {
				i = nil
				break
			}

			// Otherwise, continue to the next node.
			i = n.next
			continue
		}

		// If we didn't match the whole of the key, return.
		// No match.
		if c != len(n.nodekey) {
			return
		}

		// Exact match? If it's a value node, return it.
		// Otherwise, no match exists.
		if c == len(remaining) {
			if n.key != "" {
				value = n.value
			}
			return
		}

		// Otherwise, cut the start off remaining, and restart with
		// this node's children.
		remaining = remaining[c:]
		i = n.down
	}

	// Failed to find anything, return.
	return
}


// IterSub gets an iterator for part of the trie, which will iterate all keys
// with a given prefix. The value of the chain up to subtrie returned may be
// longer than the prefix given, if no other sub entries exist.
// Returns nil if there are no entries with the given prefix.
func (t *Trie) GetSub(prefix string) *Iterator {
	it := new(Iterator)
	it.parents = make([]*nodeBox, 0)

	remaining := prefix
	i := t.first
	for i != nil {
		n := i.n

		// Find out how much of the node key matches ours.
		var c int
		for c = 0; c < len(n.nodekey) && c < len(remaining) &&
			n.nodekey[c] == remaining[c]; c++ {
		}

		// If this key has nothing in common with ours, next.
		if c == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.next == nil {
				i = nil
				break
			}

			// Otherwise, continue to the next node.
			i = n.next
			continue
		}

		// Either is our prefix, or fully contains it?
		// This is it. If we don't already have a value, iterate it.
		// If we don't find a value, no match, otherwise return it.
		if c == len(remaining) {
			it.it = i
			if it.it.n.key != "" || it.Next() {
				return it
			}
			return nil
		}

		// Otherwise, if we didn't match the whole of the key,
		// return. No match.
		if c != len(n.nodekey) {
			return nil
		}

		// Otherwise, cut the start off remaining, and restart with
		// this node's children.
		remaining = remaining[c:]
		i = n.down
	}

	// Failed to find anything, return.
	return nil
}


// Insert adds the given key and value to the trie.
// If it already exists, it is overwritten with the new value.
// Returns the previous value of the given key, or nil if it had none.
func (t *Trie) Insert(key string, value interface{}) (old interface{}) {
	first := t.first

	// remaining contains the unmatched part of the key.
	remaining := key

	// parent points to the parent of the current node.
	var parent *nodeBox

	// i is the current sibling.
	i := first

	for i != nil {
		n := i.n

		// Find out how much of the node key matches ours.
		var c int
		for c = 0; c < len(n.nodekey) && c < len(remaining) &&
			n.nodekey[c] == remaining[c]; c++ {
		}

		// If this key has nothing in common with ours, next.
		if c == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.next == nil {
				i = nil
				break
			}

			// Otherwise, continue to the next node.
			i = n.next
			continue
		}

		// If we matched the whole of the key...
		if c == len(n.nodekey) {

			// Exact match? Make it a data node if it wasn't
			// already, and set our value.
			if c == len(remaining) {
				old = n.value
				n.value = value
				n.key = key
				return
			}

			// Otherwise, set this as our new root, cut the start
			// off remaining, and restart.
			remaining = remaining[c:]
			parent = i
			i = n.down
			continue
		}

		// Otherwise, our key matches part of this node's key, but not
		// all, and we need to split this node into two.
		newn := new(node)
		*newn = *n
		newn.nodekey = n.nodekey[c:]
		newn.next = nil
		newi := new(nodeBox)
		newi.n = newn

		newn = new(node)
		*newn = *n
		newn.down = newi
		newn.nodekey = n.nodekey[:c]

		// If none of our name is left, then the first parent node
		// above matches it. Otherwise, we need to add a second child
		// for our value.
		remaining = remaining[c:]
		if len(remaining) == 0 {
			newn.key = key
			newn.value = value
		} else {
			newn.key = ""
			newn.value = interface{}(nil)

			newchild := new(node)
			newchild.next = newn.down
			newchild.nodekey = remaining
			newchild.key = key
			newchild.value = value
			newi = new(nodeBox)
			newi.n = newchild
			newn.down = newi
		}

		// Switch the old n with the one with the new children.
		i.n = newn

		return
	}

	// If we've gone through the list and not found anything with
	// a common prefix, we need to simply add ourselves.
	if i == nil {
		i := new(nodeBox)
		n := new(node)
		n.nodekey = remaining
		n.key = key
		n.value = value
		i.n = n

		if parent != nil {
			n.next = parent.n.down
			parent.n.down = i
		} else {
			n.next = first
			t.first = i
		}
	}
	return
}


// Remove deletes a key from the trie.
// If it does not exist, nothing happens.
// Returns the value of the removed key, or nil if it did not exist.
func (t *Trie) Remove(key string) (old interface{}) {
	var prev *nodeBox

	parents := make([]*nodeBox, 0)
	remaining := key
	i := t.first
	for i != nil {
		n := i.n

		// Find out how much of the node key matches ours.
		var c int
		for c = 0; c < len(n.nodekey) && c < len(remaining) &&
			n.nodekey[c] == remaining[c]; c++ {
		}

		// If this key has nothing in common with ours, next.
		if c == 0 {
			// If this is the last sibling, we're out of
			// nodes to check. No match.
			if n.next == nil {
				i = nil
				break
			}

			// Otherwise, continue to the next node.
			prev = i
			i = n.next
			continue
		}

		// If we didn't match the whole of the key, return.
		// No match.
		if c != len(n.nodekey) {
			return
		}

		// Exact match? If it's a value node, delete it.
		// Otherwise, no match exists.
		if c == len(remaining) {

			// If it's not a value node, no match.
			if n.key == "" {
				return
			}

			old = n.value

			// If the node has children, we can't delete it
			// without breaking concurrent access.
			if n.down != nil {
				newn := new(node)
				*newn = *n
				newn.key = ""
				newn.value = interface{}(nil)

				i.n = newn
				return
			}

			// If we have a previous node, remove us from it.
			if prev != nil {
				prev.n.next = n.next
				return
			}

			// If we have no parent either, we're the first node.
			if len(parents) == 0 {
				t.first = n.next
				return
			}

			// Otherwise, remove us from our parent, and 
			// recursively delete the parent if they have no value
			// or remaining children. We have to find their
			// previous node to do this.
			for len(parents) > 0 {
				parent := parents[len(parents)-1]
				parents = parents[:len(parents)-1]
				
				prev := &parent.n.down
				for *prev != i && *prev != nil {
					prev = &((*prev).n.next)
				}

				if *prev != nil {
					*prev = i.n.next
				} else {
					return
				}

				i = parent
				if i.n.down != nil || i.n.key != "" {
					return
				}
			}

			return
		}

		// Otherwise, set this as our new parent,
		// cut the start off remaining, and restart.
		parents = append(parents, i)
		remaining = remaining[c:]
		prev = nil
		i = n.down
	}

	// Failed to find anything, return.
	return
}
