/*
	Provides an implementation of an iterable radix tree.

	Designed for storing small numbers of items with minimal overhead,
	and large numbers with acceptable overhead, and to be iterable without
	invalidating iterators in any operation, including over subsets of the
	contents.

	This container is not threadsafe; external synchronisation mechanisms
	should be used. Variations on the container are provided for storing
	different types as values; their names indicate the values they store.
*/
package trie
