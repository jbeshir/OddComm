/*
	Provides an implementation of a concurrency-safe iterable radix tree.

	Designed for storing small numbers of items with minimal overhead,
	and large numbers with acceptable overhead, and to be iterable without
	invalidating iterators in any operation, including over subsets of the
	contents.

	This container can be concurrently used by any number of goroutines
	simultaneously, and is lock-free for reading and inserting values.
*/
package trie
