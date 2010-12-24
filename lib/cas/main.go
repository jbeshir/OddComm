/*
	Cas provides a function for the compare-and-swap atomic operation,
	permitting atomic writes to structures with concurrent access.

	This package simply provides a slightly modified version of an ASM
	function taken from the Go sync	package.
*/
package cas

import "unsafe"


func Cas(val unsafe.Pointer, old, new unsafe.Pointer) bool
