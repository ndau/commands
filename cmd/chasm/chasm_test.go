package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"testing"
)

func TestSimple1(t *testing.T) {
	code := `
		; comment
		func foo(1) {
		nop
		}
`
	checkParse(t, "Simple1", code, "800001 00 88")
}

func TestSimple2(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			nop ; nop instruction
			drop ; drop nothing
		}
`
	checkParse(t, "Simple2", code, "800001 0001 88")
}

func TestSimplePush(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			push 0
		}
`
	checkParse(t, "SimplePush", code, "800001 20 88")
}

func TestNumberFormats(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			push 0xabcd
			push 01777
			push 0b1010101001010101
			push 0b001_1101
			push 0xbad_cab
			push 127
		}
`
	checkParse(t, "NumberFormats", code, "800001 23cdab00 22ff03 2355aa00 211d 24abdcba00 217f 88")
}

func TestPushB(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			pushb 5 1 2 3 4 5 6 7 8 9 10
			pushb "HI!"
			pushb 0x01 0x02 0x03
			pushb 0x01 0x02 0x03 0x04 0x05 0x06 0x07
			pushb addr(deadbeeffeed5d00d5)
		}
`
	checkParse(t, "PushB", code, `800001
	2A 0b 050102030405060708090a
	2A 03 484921
	2A 03 01 02 03
	2A 07 01 02 03 04 05 06 07
	2A 09 de ad be ef fe ed 5d 00 d5
	88`)
}

func TestPushT(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			pusht 2018-02-03T01:23:45Z
		}
`
	checkParse(t, "PushT", code, `800001
	2b 40 6a e1 72  43 07 02 00
	88`)
}

func TestFunc(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			zero
			call bar
		}
		func buzz(0) {
		}
		func bar(2) {
			one
			add
		}
`
	checkParse(t, "Func", code, "800001 20 8102 88 800100 88 800202 1a 40 88")
}

func TestSeveralPushes(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			push -1
			push 1
			push 2
			push 12
		}
`
	checkParse(t, "SeveralPushes", code, "800001 1b1a2102210c 88")
}

func TestConstants(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			K = 65535
			push K
		}
`
	checkParse(t, "Constants", code, "800001 23FFFF00 88")
}

func TestUnitaryOpcodes1(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			nop
			drop
			drop2
			dup
			dup2
			swap
			over
			ret
			fail
			zero
			false
			one
			true
			neg1
			maxnum
			minnum
			now
			rand
			add
			sub
			mul
			div
			mod
			divmod
			muldiv
			not
			neg
			inc
			dec
			index
			len
			append
			extend
			slice
		}
`
	checkParse(t, "Unitary1", code, `
		800001
		00 0102 0506 090C
		1011 2020 1a 1b1b 1c1d 2c
		2e40 4142 4344 4546
		4849 4A4B 5051 5253
		5488`)
}

func TestUnitaryOpcodes2(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			choice
			ifz
			ifnz
			else
			endif
			sum
			avg
			max
			min
			pushl
			lt
			lte
			eq
			gte
			gt
		}
`
	checkParse(t, "Unitary2", code,
		"80000194 898a8e8f909192932f c0c1c2c3c4 88")
}

func TestBinary(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			pick 2
			pick 12
			roll 0xA
			field 3
			fieldl 0
			call bar
			lookup bar
			deco bar 7
			sort 3
			wchoice 1
		}
		func bar(0) {
			nop
		}
`
	checkParse(t, "Binary", code, "800001 0D020D0C0E0A 6003 7000 8101 9701 820107 9603 9501 88 800100 00 88")
}

func TestRealistic(t *testing.T) {
	code := `
		; This program pushes a, b, c,
		; and x on the stack and calculates
		; a*x*x + b*x + c
		handler 0 {
			A = 3
			B = 5
			C = 7
			X = 21

			push A
			push B
			push C
			push X	; ABCX
			roll 4	; BCXA
			pick 1	; BCXAX
			dup  	; BCXAXX
			mul		; BCXAR
			mul		; BCXR
			roll 4  ; CXRB
			roll 2  ; CRBX
			mul		; CRS
			add		; CR
			add		; R
			ret
		}
`
	checkParse(t, "Realistic", code, `
		a0 00 21 03 21 05 21 07 21  15 0E 04 0d 01 05 42 42
		0e 04 0e 02 42 40 40 10 88`)
}

func TestIfZNZ(t *testing.T) {
	code := `
		; comment
		func foo(1) {
			nop
			ifz
			ifnz
		}
`
	checkParse(t, "Simple1", code, "800001 00898a 88")
}

func TestCallFromEvent(t *testing.T) {
	code := `
		handler 0 {
			one
			call double
			push 2
			eq
			not                             ; invert meaning for return code
		}

		func justzero(1) {
			zero
		}

		func double(1) {
			dup
			add
		}
	`
	checkParse(t, "CallFromEvent", code, "a000 1a 8101 2102 c2 48 88  800001 20 88  800101 05 40 88")
	// expected miscompilation:          "a000 1a 8100 2102 c2 48 88  800001 20 88  800101 05 40 88"
	//                                               ^
	// at present, it is impossible to call any function except the first from
	// within an event handler. This test will let us know when we have fixed that.
}

func TestConstantTimestamp(t *testing.T) {
	code := `
		GENESIS = 2019-05-11T03:46:40.570549Z
		handler 0 {
			pusht GENESIS
		}
	`
	checkParse(t, "ConstantTimestamp", code, "a000 2b b57cb54c 932b0200 88")
}
