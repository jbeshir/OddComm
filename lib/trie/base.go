package trie

// If this file is not called base.go, it is an automatically generated
// variation on it, in which case you should not edit directly nor commit it.

// Iterator safety: Thanks to garbage collection, we get 90% of this for free.
// If we have a pointer to a node, then all its pointers will still be valid.
// Eventually, we will return to a part of the trie that still exists, and,
// if no one else is pointing to the nodes we iterated, they will be deleted.
// The only gotcha is that we cannot change the parent or grandparent node,
// as this would cause us to miss our 'root' node while iterating a subtrie.
// Trie contains a single trie, 


// Trie provides the implementation of the radix trie.
// It contains methods on the trie and a pointer to its first node.
// Individual nodes are accessed by generating a subtrie starting at them.
// An empty trie uses no more space than a single nil pointer.
type Trie struct {
	First *trieNode
}

// trieNode is a single node of a Trie.
type trieNode struct {
	down *trieNode // First child of this node.
	next *trieNode // Next sibling of this node- or the next one back
	// up if it has none left.
	nodekey string      // Value of this node.
	key     string      // Full key of this node, if it has a key.
	value   interface{} // If set, this node contains a value.
	lastSib bool        // Whether this node is the last of its siblings.
}


// Value returns this trie node's key and value, or nil if it is not a value
// node.
func (t *Trie) Value() (name string, value interface{}) {
	if t.First != nil {
		name, value = t.First.key, t.First.value
	}
	return
}


// Next gets the next value node, not including this node itself if it has
// a value. If end is non-empty, it indicates a root node to not go past,
// used to iterate only within a subtrie.
//
// Calling Next on a nil (that is, empty) trie is bad.
func (t *Trie) Next(end Trie) (next Trie) {

	next.First = next.First.nextNode(end.First)
	for next.First != end.First && next.First != nil && next.First.key == "" {
		next.First = next.First.nextNode(end.First)
	}

	// If we found the end node, return nil, not it.
	if next.First == end.First {
		next.First = nil
	}

	return
}

// Iterate to the next node, including non-value nodes. May return nil.
func (t *trieNode) nextNode(end *trieNode) *trieNode {
	if t.down != nil {
		t = t.down
	} else {
		for cont := true; cont && t != nil; {

			// If this is the end, we do not want to go to its
			// next node, so return nil.
			if t == end {
				return nil
			}

			// If this was a last sibling, we're going back up to
			// its parent; continue onto ITS next node, not
			// recursing into children.
			cont = false
			if t.lastSib {
				cont = true
			}

			t = t.next
		}
	}
	return t
}


// Empty returns whether this trie is empty.
// Used to see if a result from Next() is still a valid trie.
func (t *Trie) Empty() bool {
	return t.First == nil
}


// Add adds the given key and value to the trie.
// If it already exists, it is overwritten with the new value.
// Returns the previous value of the given key, or nil if it had none.
func (t *Trie) Add(key string, value interface{}) (old interface{}) {

	// remaining contains the unmatched part of the key.
	remaining := key

	// parent tracks the parent node to the current sibling list.
	// Used to add a node.
	var parent *trieNode

	// n is the current sibling.
	n := t.First

	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
			n.nodekey[i] == remaining[i]; i++ {
		}

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
			remaining = remaining[i:]
			parent = n
			n = n.down
			continue
		}

		// Otherwise, our key matches part of this node's key, but not
		// all, and we need to split this node into two.
		newn := new(trieNode)
		newn.down = n.down
		newn.next = n
		newn.nodekey = n.nodekey[i:]
		newn.lastSib = true
		newn.key = n.key
		newn.value = n.value
		n.down = newn
		n.nodekey = n.nodekey[0:i]
		n.key = ""
		n.value = interface{}(nil)

		// If none of our name is left, then the first parent node
		// above matches it. Otherwise, we need to add ourselves as a
		// new child.
		if len(remaining) == i {
			newn = n
		} else {
			newn = new(trieNode)
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
		n := new(trieNode)
		n.nodekey = remaining
		n.key = key
		n.value = value

		// Set the next node, which is the first sibling. If there
		// isn't one, it's the parent, and we're the last sibling.
		if parent != nil {
			n.next = parent.down
		}
		if n.next == nil {
			n.next = parent
			n.lastSib = true
		}

		// If we have a parent, set this node as its child.
		// If not, set this as the root node.
		if parent != nil {
			parent.down = n
		} else {
			t.First = n
		}
	}
	return
}


// Del deletes a key from the trie.
// If it does not exist, nothing happens.
// Returns the value of the removed key, or nil if it did not exist.
func (t *Trie) Del(key string) (old interface{}) {
	remaining := key
	root := t.First
	n := root
	var prev *trieNode
	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
			n.nodekey[i] == remaining[i]; i++ {
		}

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

				// If the node has children, we can't delete it
				// without breaking iterators. However, if it
				// has only one child and that child has no
				// children, we can merge with that child.
				if n.down != nil {
					c := n.down
					if c.next != n || c.down != nil {
						n.key = ""
						n.value = interface{}(nil)
						return
					}

					// Merge our one child into us.
					// This HAS to happen this way around,
					// or we'd be editing the child's
					// parent (breaks iterators).
					n.nodekey = n.nodekey + n.down.nodekey
					n.key = n.down.key
					n.value = n.down.value
					n.down = nil
					return
				}

				// We have no children, so just delete
				// references to this node.
				if prev != nil {
					prev.next = n.next
					prev.lastSib = n.lastSib
				} else {
					// If we had siblings, just remove us.
					if !n.lastSib {
						root = n.next
						return
					}

					// Otherwise, recurse up, deleting
					// parents without children or values.
					n.key = ""
					n.value = interface{}(nil)
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
							if t.First == n {
								t.First = nil
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
		root = n.down
		remaining = remaining[i:]
		prev = nil
		n = root
	}

	// Failed to find anything, return.
	return
}


// Get gets a key's value from the trie.
// If it does not exist, returns nil.
func (t *Trie) Get(key string) (value interface{}) {
	remaining := key
	n := t.First
	for n != nil {

		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
			n.nodekey[i] == remaining[i]; i++ {
		}

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

		// Otherwise, cut the start off remaining, and restart with
		// this node's children.
		remaining = remaining[i:]
		n = n.down
	}

	// Failed to find anything, return.
	return
}

// GetSub gets a subtrie from the trie, consisting of a root of all values
// with a given prefix. The value of the chain up to subtrie returned may be
// longer than the prefix given, if no other sub entries exist.
// The only safe operation on a subtrie is iterating it with Next(). The Trie*
// functions require that they be working on keys from the start of the tree.
// Returns nil if there are no entries with the given prefix.
func (t *Trie) GetSub(prefix string) (subtree Trie) {
	remaining := prefix
	n := t.First
	for n != nil {
		// Find out how much of the node key matches ours.
		var i int
		for i = 0; i < len(n.nodekey) && i < len(remaining) &&
			n.nodekey[i] == remaining[i]; i++ {
		}

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
			subtree.First = n
			return
		}

		// Otherwise, if we didn't match the whole of the key,
		// return. No match.
		if i != len(n.nodekey) {
			return
		}

		// Otherwise, cut the start off remaining, and restart with
		// this node's children.
		remaining = remaining[i:]
		n = n.down
	}

	// Failed to find anything, return.
	return
}