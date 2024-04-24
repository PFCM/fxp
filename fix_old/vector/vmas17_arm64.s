// +build !noasm,!gccgo,!safe

#include "textflag.h"

// func _vmas17_serial(a, b, c, d []fix.S17, n int)
TEXT ·_vmas17_serial(SB), NOSPLIT, $0
	MOVD a+0(FP), R0   // addr of first element in a
	MOVD b+24(FP), R1  // addr of first element b
	MOVD c+48(FP), R2  // addr of first element in c
	MOVD d+72(FP), R3  // addr of first element in d
	MOVD n+96(FP), R4  // r4 = n
	MOVD ZR, R5        // R5 = 0
serial_compare:
	CMP   R5, R4
	BLE   serial_end       // jump to end if r5 >= r4
	VLD1 R0, [V0.B16]
	VMOVQ (R1)(R5<<4), V1  
	VMOVQ (R2)(R5<<4), V2  
	;; FMADDD F1, F2, F0, F0	// F0 = F0 + F1
	;; FMOVD F0, (R3)(R5<<3)  // d[R5] = F0
	ADD   $16, R5, R5       // R5 ++
	JMP   serial_compare
serial_end:
	RET

