package trie

import "sync"
import "unsafe"

import "oddcomm/lib/cas"


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
// An empty trie uses no more space than a nil pointer and an int32.
type Trie struct {
	first *nodeBox      // First node pointer.
	rem   sync.Mutex // Removal mutex.
}

// nodeBox contains a single trie node. It wraps the node, so it can be changed
// atomically without invalidating pointers.
type nodeBox struct {
	n *node
}

// node stores the contents of a trie node.
type node struct {
	nodekey string      // Key of this node, not including parent nodekeys.
	down    *nodeBox       // First child of this node (nilable).
	next    *nodeBox       // Next sibling of this node (nilable).
	key     string      // Full key of this node, if it has a key.
	value   interface{} // If set, this node contains a value.
}


// Iterator permits iteration of a part of a trie. It keeps a stack of visited
// parent nodes to return to, and will iterate only until it reaches the end
// of the trie or returns upwards to the same level it was called on.
// A single iterator may not be used concurrently, but separate iterators may.
type Iterator struct {
	parents []*nodeBox
	it *nodeBox
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
			if n.next == nil {
				if len(i.parents) > 1 {
					i.parents = i.parents[0:len(i.parents)-1]
					i.it = i.parents[len(i.parents)-1]
				} else {
					i.it = nil
				}
			} else {
				i.it = n.next
				break
			}
			if i.it != nil {
				n = i.it.n
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

// If the trie is concurrently changed while this function is running, we
// restart.
spin:
	first := t.first

	// remaining contains the unmatched part of the key.
	remaining := key

	// parent points to the parent of the current node.
	var parent *nodeBox
	var pn *node

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
				newn := new(node)
				*newn = *n
				newn.key = key
				newn.value = value
				if !cas.Cas((unsafe.Pointer)(&(i.n)), unsafe.Pointer(n), unsafe.Pointer(newn)) {
					// You spin me right round baby right round
					goto spin
				}
				return
			}

			// Otherwise, set this as our new root, cut the start
			// off remaining, and restart.
			remaining = remaining[c:]
			parent = i
			pn = n
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
		if !cas.Cas(unsafe.Pointer(&(i.n)), unsafe.Pointer(n), unsafe.Pointer(newn)) {
			// Like a record baby right round round
			goto spin
		}

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
			n.next = pn.down

			newn := new(node)
			*newn = *pn
			newn.down = i

			// Switch the old parent node with the new one.
			if !cas.Cas(unsafe.Pointer(&(parent.n)), unsafe.Pointer(pn), unsafe.Pointer(newn)) {
        			goto spin
			}
		} else {
			n.next = first

			// We're the first node! Swap a pointer to us with that.
			if !cas.Cas(unsafe.Pointer(&(t.first)), unsafe.Pointer(first), unsafe.Pointer(i)) {
        			goto spin
			}
		}
	}
	return
}


// Remove deletes a key from the trie.
// If it does not exist, nothing happens.
// Returns the value of the removed key, or nil if it did not exist.
func (t *Trie) Remove(key string) (old interface{}) {

spin:
	parents := make([]*nodeBox, 1)
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

		// Exact match? If it's a value node, delete it.
		// Otherwise, no match exists.
		if c == len(remaining) {
			if n.key != "" {
				old = n.value

				// If the node has children, we can't delete it
				// without breaking concurrent access.
				if n.down != nil {
					newn := new(node)
					*newn = *n
					newn.key = ""
					newn.value = interface{}(nil)

					if !cas.Cas(unsafe.Pointer(&(i.n)), unsafe.Pointer(n), unsafe.Pointer(newn)) {
						goto spin
					}
					return
				}

				// We have no children, so commence deletion;
				// set the node's nodekey to "".
				newn := new(node)
				*newn = *n
				newn.key = ""
				newn.value = interface{}(nil)
				newn.nodekey = ""
				if !cas.Cas(unsafe.Pointer(&(i.n)), unsafe.Pointer(n), unsafe.Pointer(newn)) {
					goto spin
				}

				// Lock the removal mutex.
				t.rem.Lock()

				// Get the node's parent.
				var parent *nodeBox
				if len(parents) > 0 {
					parent = parents[len(parents)-1]
					parents = parents[:len(parents)-1]
				}

				// Find and delete this node.
				parent, ok := t.removeNode(parent, i)
				if !ok {
					// Someone else got here first!
					t.rem.Unlock()
					return
				}

				// While we deleted from a parent...
				for parent != nil {
					pn := parent.n

					// If the parent has a value or
					// children, leave it alone.
					if pn.down != nil || pn.key != "" {
						break
					}

					// Set its nodekey to "".
					newn := new(node)
					*newn = *pn
					newn.key = ""
					newn.value = interface{}(nil)
					newn.nodekey = ""
					if !cas.Cas(unsafe.Pointer(&(parent.n)), unsafe.Pointer(pn), unsafe.Pointer(newn)) {
						// Stop; parent changed.
						break
					}
					
					// Get its parent.
					newparent := (*nodeBox)(nil)
					if len(parents) > 0 {
						newparent = parents[len(parents)-1]
						parents = parents[:len(parents)-1]
					}

					// Find and delete it.
					parent, ok = t.removeNode(newparent, parent)
					if !ok {
						// Already gone?
						break
					}
				}

				// Release the removal mutex.
				t.rem.Unlock()
			}
			return
		}

		// Otherwise, set this as our new parent,
		// cut the start off remaining, and restart.
		remaining = remaining[c:]
		i = n.down
	}

	// Failed to find anything, return.
	return
}


// removeNode deletes the given node from the Trie.
// The node must be marked for deletion (nodekey set to "") and the caller must
// hold the removal mutex, or this is terribly unsafe.
//
// parent is nil, unless the node was its parents' first and only child, in
// which case it is the node's parent, for consideration for deletion itself.
// ok indicates whether the deletion was successful.
func (t *Trie) removeNode(p, target *nodeBox) (parent *nodeBox, ok bool) {

spin:
	it := new(Iterator)
	it.parents = make([]*nodeBox, 1)

	// If they have a parent, check if they are the first child.
	if p != nil {
		pn := p.n
		if pn.down == target {
			newn := new(node)
			*newn = *pn
			newn.down = target.n.next
			if !cas.Cas(unsafe.Pointer(&(p.n)), unsafe.Pointer(pn), unsafe.Pointer(newn)) {
				goto spin
			}
			if newn.down == nil {
				parent = p
			}
			return parent, true
		}
		it.it = pn.down
	} else {
		// Handle deleting the first node.
		first := t.first
		if first == target {
			if !cas.Cas(unsafe.Pointer(&(t.first)), unsafe.Pointer(target), unsafe.Pointer(target.n.next)) {
				goto spin
			}
			return nil, true
		}
		it.it = first
	}

	// Otherwise, simply simply iterate through the siblings until we find
	// the one before the node.	
	for it.Next() {
		n := it.it.n
		if n.next == target {
			newn := new(node)
			*newn = *n
			newn.next = target.n.next
			if !cas.Cas(unsafe.Pointer(&(it.it.n)), unsafe.Pointer(n), unsafe.Pointer(newn)) {
				goto spin
			}
			return nil, true
		}
	}

	// Not found the node. Someone else must have got to it first.
	return nil, false
}
