// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// func cas(val *unsafe.Pointer, old, new unsafe.Pointer) bool
// Atomically:
//	if *val == old {
//		*val = new;
//		return true;
//	}else
//		return false;
TEXT ·Cas(SB), 7, $0
	MOVQ	8(SP), BX
	MOVQ	16(SP), AX
	MOVQ	24(SP), CX
	LOCK
	CMPXCHGQ	CX, 0(BX)
	JZ ok
	MOVL	$0, 32(SP)
	RET
ok:
	MOVL	$1, 32(SP)
	RET
