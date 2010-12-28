/*
	Provides an implementation of a iterable radix tree.

	Designed for storing small numbers of items with minimal overhead,
	and large numbers with acceptable overhead, and to be iterable without
	invalidating iterators in any operation, including over subsets of the
	contents.

	This container can be concurrently read from by any number of
	goroutines simultaneously, even while being written to, but can only
	be written to by one goroutine at once.
*/
package trie
