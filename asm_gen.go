// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2005-2016, Wojciech Muła. All rights reserved.
// Use of this source code is governed by a
// Simplified BSD License license that can be found in
// the LICENSE file.

// +build ignore

package main

import (
	"bytes"

	"github.com/tmthrgd/asm"
)

const encodeHeader = `// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2005-2016, Wojciech Muła. All rights reserved.
// Use of this source code is governed by a
// Simplified BSD License license that can be found in
// the LICENSE file.
//
// This file is auto-generated - do not modify

// +build amd64,!gccgo,!appengine
`

const decodeHeader = `// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// This file is auto-generated - do not modify

// +build amd64,!gccgo,!appengine
`

func repeat(b byte, l int) []byte {
	return bytes.Repeat([]byte{b}, l)
}

type encode struct {
	*asm.Asm

	di, si, cx asm.Register

	ret, tail asm.Label

	compare25, compare51, compare62, compare63 asm.Operand

	baseUpper, baseLower, baseZero, base62, base63 asm.Operand

	shuf, shufOut, and asm.Data
}

func (e *encode) vpand_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pand(ops[0], ops[2])
}

func (e *encode) vpcmpgtb_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pcmpgtb(ops[0], ops[2])
}

func (e *encode) vpcmpeqb_sse(ops ...asm.Operand) {
	if len(ops) != 3 {
		panic("wrong number of operands")
	}

	if ops[1] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pcmpeqb(ops[0], ops[1])
}

func (e *encode) vpblendvb_sse(ops ...asm.Operand) {
	if len(ops) != 4 {
		panic("wrong number of operands")
	}

	if ops[0] == asm.X0 || ops[0] == ops[2] {
		panic("invalid register choice fallback")
	}

	if ops[3] != asm.X0 {
		e.Movou(asm.X0, ops[3])
	}

	if ops[0] != ops[1] {
		e.Movou(ops[0], ops[1])
	}

	e.Pblendvb(ops[0], ops[2], asm.X0)
}

func (e *encode) Unpack(vpand func(ops ...asm.Operand)) {
	e.Pshufb(asm.X1, e.shuf)

	vpand(asm.X0, asm.X1, e.and.Offset(0))

	e.Pslll(asm.X1, asm.Constant(4))
	e.Pand(asm.X1, e.and.Offset(16))

	e.Por(asm.X1, asm.X0)

	vpand(asm.X0, asm.X1, e.and.Offset(32))

	e.Pslll(asm.X1, asm.Constant(2))
	e.Pand(asm.X1, e.and.Offset(48))

	e.Por(asm.X1, asm.X0)

	e.Pshufb(asm.X1, e.shufOut)
}

func (e *encode) Lookup(vpcmpgtb, vpcmpeqb, vpblendvb func(ops ...asm.Operand)) {
	vpcmpgtb(asm.X2, asm.X1, e.compare25)
	vpcmpgtb(asm.X3, asm.X1, e.compare51)
	vpcmpeqb(asm.X4, asm.X1, e.compare62)
	vpcmpeqb(asm.X5, asm.X1, e.compare63)

	vpblendvb(asm.X2, e.baseUpper, e.baseLower, asm.X2)
	vpblendvb(asm.X2, asm.X2, e.baseZero, asm.X3)

	e.Paddb(asm.X1, asm.X2)

	vpblendvb(asm.X1, asm.X1, e.base62, asm.X4)
	vpblendvb(asm.X1, asm.X1, e.base63, asm.X5)
}

func (e *encode) Convert(vpand, vpcmpgtb, vpcmpeqb, vpblendvb func(ops ...asm.Operand)) {
	e.Unpack(vpand)
	e.Lookup(vpcmpgtb, vpcmpeqb, vpblendvb)
}

func (e *encode) BigLoop(l asm.Label, vpand, vpcmpgtb, vpcmpeqb, vpblendvb func(ops ...asm.Operand)) {
	e.Label(l)

	e.Movou(asm.X1, asm.Address(e.si))

	e.Convert(vpand, vpcmpgtb, vpcmpeqb, vpblendvb)

	e.Movou(asm.Address(e.di), asm.X1)

	e.Subq(e.cx, asm.Constant(12))
	e.Jz(e.ret)

	e.Addq(e.si, asm.Constant(12))
	e.Addq(e.di, asm.Constant(16))

	e.Cmpq(asm.Constant(16), e.cx)
	e.Jae(l)

	e.Cmpq(asm.Constant(3), e.cx)
	e.Jb(e.tail)
}

func encodeASM(a *asm.Asm) {
	shuf := a.Data32("encodeShuf", []uint32{
		0xff000102,
		0xff030405,
		0xff060708,
		0xff090a0b,
	})
	shufOut := a.Data32("encodeShufOut", []uint32{
		0x00010203,
		0x04050607,
		0x08090a0b,
		0x0c0d0e0f,
	})
	and := a.Data64("encodeAnd", []uint64{
		0x00000fff00000fff,
		0x00000fff00000fff,
		0x0fff00000fff0000,
		0x0fff00000fff0000,
		0x003f003f003f003f,
		0x003f003f003f003f,
		0x3f003f003f003f00,
		0x3f003f003f003f00,
	})
	compare := a.Data("encodeCompare", bytes.Join([][]byte{
		repeat(25, 16),
		repeat(51, 16),
		repeat(62, 16),
		repeat(63, 16),
	}, nil))
	var c52 byte = 52
	base := a.Data("encodeBase", bytes.Join([][]byte{
		repeat('A', 16),
		repeat('a' - 26, 16),
		repeat('0' - c52, 16),
		repeat('+', 16),
		repeat('/', 16),
		repeat('-', 16),
		repeat('_', 16),
	}, nil))
	lookup := a.DataString("encodeLookup",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

	a.NewFunction("encodeASM")
	a.NoSplit()

	dst := a.Argument("dst", 8)
	src := a.Argument("src", 8)
	length := a.Argument("len", 8)
	padding := a.Argument("padding", 4)
	url := a.Argument("url", 4)

	a.Start()

	bigloop_avx := a.NewLabel("bigloop_avx")
	bigloop_sse := a.NewLabel("bigloop_sse")
	loop := a.NewLabel("loop")
	loop_preheader := loop.Suffix("preheader")
	tail := a.NewLabel("tail")
	ret := a.NewLabel("ret")

	e := &encode{
		a,

		asm.DI, asm.SI, asm.BX,

		ret, tail,

		asm.Address(asm.R14, 0), asm.Address(asm.R14, 16), asm.Address(asm.R14, 32), asm.Address(asm.R14, 48),

		asm.X11, asm.X12, asm.X13, asm.X14, asm.X15,

		shuf, shufOut, and,
	}

	a.Movq(e.di, dst)
	a.Movq(e.si, src)
	a.Movq(e.cx, length)
	a.Movl(asm.AX, padding)
	a.Xorq(asm.R14, asm.R14)
	a.Movb(asm.R14, url)

	a.Shlq(asm.R14, asm.Constant(3))

	a.Movq(asm.DX, lookup.Address())
	a.Leaq(asm.DX, asm.Address(asm.DX, asm.R14, asm.SX8))

	a.Cmpq(asm.Constant(3), e.cx)
	a.Jb(tail)

	a.Cmpq(asm.Constant(16), e.cx)
	a.Jb(loop_preheader)

	a.Shrq(asm.R14, asm.Constant(1))

	a.Movou(e.baseUpper, base.Offset(0))
	a.Movou(e.baseLower, base.Offset(16))
	a.Movou(e.baseZero, base.Offset(32))
	a.Movq(asm.R13, base.Address())
	a.Movou(e.base62, asm.Address(asm.R13, asm.R14, asm.SX8, 48))
	a.Movou(e.base63, asm.Address(asm.R13, asm.R14, asm.SX8, 64))

	a.Movq(asm.R14, compare.Address())

	a.Cmpb(asm.Constant(1), asm.Data("runtime·support_avx"))
	a.Jne(bigloop_sse)

	e.BigLoop(bigloop_avx, a.Vpand, a.Vpcmpgtb, a.Vpcmpeqb, a.Vpblendvb)

	a.Label(loop_preheader)
	a.Xorq(asm.R9, asm.R9)
	a.Xorq(asm.R10, asm.R10)
	a.Xorq(asm.R11, asm.R11)

	a.Label(loop)

	for i, r := range []asm.Operand{asm.R9, asm.R10, asm.R11} {
		a.Movb(r, asm.Address(e.si, 2 - i))
	}

	a.Movq(asm.R12, asm.R9)
	a.Andb(asm.R12, asm.Constant(0x3f))

	a.Movq(asm.R13, asm.R10)
	a.Andb(asm.R13, asm.Constant(0x0f))
	a.Shlb(asm.R13, asm.Constant(2))
	a.Shrb(asm.R9, asm.Constant(6))
	a.Orb(asm.R13, asm.R9)

	a.Movq(asm.R14, asm.R11)
	a.Andb(asm.R14, asm.Constant(0x03))
	a.Shlb(asm.R14, asm.Constant(4))
	a.Shrb(asm.R10, asm.Constant(4))
	a.Orb(asm.R14, asm.R10)

	a.Shrb(asm.R11, asm.Constant(2))

	for _, r := range [][]asm.Operand{
		{asm.R12, asm.R12},
		{asm.R13, asm.R13},
		{asm.R14, asm.R14},
		{asm.R11, asm.R15},
	} {
		a.Movb(r[1], asm.Address(asm.DX, r[0], asm.SX1))
	}

	for i, r := range []asm.Operand{asm.R12, asm.R13, asm.R14, asm.R15} {
		a.Movb(asm.Address(e.di, 3 - i), r)
	}

	a.Subq(e.cx, asm.Constant(3))
	a.Jz(ret)

	a.Addq(e.si, asm.Constant(3))
	a.Addq(e.di, asm.Constant(4))

	a.Cmpq(asm.Constant(3), e.cx)
	a.Jae(loop)

	a.Label(tail)

	tail1 := tail.Suffix("1")

	a.Xorq(asm.R11, asm.R11)
	a.Movb(asm.R11, asm.Address(e.si))

	a.Movq(asm.R14, asm.R11)
	a.Andb(asm.R14, asm.Constant(0x03))
	a.Shlb(asm.R14, asm.Constant(4))

	a.Cmpq(asm.Constant(2), e.cx)
	a.Jb(tail1)

	a.Xorq(asm.R10, asm.R10)
	a.Movb(asm.R10, asm.Address(e.si, 1))

	a.Movb(asm.R9, asm.R10)
	a.Shrb(asm.R9, asm.Constant(4))
	a.Orb(asm.R14, asm.R9)

	a.Andb(asm.R10, asm.Constant(0x0f))
	a.Shlb(asm.R10, asm.Constant(2))

	a.Movb(asm.R13, asm.Address(asm.DX, asm.R10, asm.SX1))
	a.Movb(asm.Address(e.di, 2), asm.R13)

	a.Label(tail1)

	a.Shrb(asm.R11, asm.Constant(2))

	a.Movb(asm.R14, asm.Address(asm.DX, asm.R14, asm.SX1))
	a.Movb(asm.R15, asm.Address(asm.DX, asm.R11, asm.SX1))

	a.Movb(asm.Address(e.di, 1), asm.R14)
	a.Movb(asm.Address(e.di, 0), asm.R15)

	a.Cmpl(asm.Constant(-1), asm.AX)
	a.Je(ret)

	a.Movb(asm.Address(e.di, e.cx, asm.SX1, 1), asm.AX)

	a.Cmpq(asm.Constant(2), e.cx)
	a.Je(ret)

	a.Movb(asm.Address(e.di, e.cx, asm.SX1, 2), asm.AX)

	a.Label(ret)
	a.Ret()

	e.BigLoop(bigloop_sse, e.vpand_sse, e.vpcmpgtb_sse, e.vpcmpeqb_sse, e.vpblendvb_sse)
	a.Jmp(loop_preheader)
}

type decode struct {
	*asm.Asm

	di, si, cx asm.Register

	ret, invalid asm.Label

	lowerBound, upperBound asm.Operand

	shifts asm.Operand

	nibble asm.Operand

	x2f, x2fOffset asm.Operand

	mergeBytes, mergeWords asm.Operand
}

func (d *decode) Convert() {
	d.Vpsrld(asm.X2, asm.X1, asm.Constant(4))
	d.Pand(asm.X2, d.nibble)

	d.Vpshufb(asm.X3, d.lowerBound, asm.X2)
	d.Vpshufb(asm.X0, d.upperBound, asm.X2)

	d.Vpcmpgtb(asm.X4, asm.X3, asm.X1)
	d.Vpcmpgtb(asm.X5, asm.X1, asm.X0)
	d.Vpcmpeqb(asm.X6, asm.X1, d.x2f)

	d.Por(asm.X4, asm.X5)
	d.Pandn(asm.X6, asm.X4)

	d.Pmovmskb(asm.AX, asm.X6)

	/*d.Testw(asm.DX, asm.AX)
	d.Jnz(d.invalid)*/

	d.Vpshufb(asm.X7, d.shifts, asm.X2)

	d.Psubb(asm.X1, asm.X0)
	d.Paddb(asm.X1, asm.X7)

	d.Pand(asm.X6, d.x2fOffset)
	d.Paddb(asm.X1, asm.X7)

	d.Pmaddubsw(asm.X1, d.mergeBytes)
	d.Pmaddwl(asm.X1, d.mergeWords)
}

func (d *decode) BigLoop(l asm.Label) {
	d.Label(l)

	d.Movou(asm.X1, asm.Address(d.si))

	d.Convert()

	d.Movou(asm.Address(d.di), asm.X1)

	d.Subq(d.cx, asm.Constant(16))
	d.Jz(d.ret)

	d.Addq(d.si, asm.Constant(16))
	d.Addq(d.di, asm.Constant(12))

	return

	d.Cmpq(asm.Constant(16+8), d.cx)
	d.Jae(l)
}

func decodeASM(a *asm.Asm) {
	var linv, hinv byte = 1, 0
	lowerBound := a.Data("decodeLowerBound", []byte{
		linv, linv, 0x2b, 0x30,
		0x40, 0x50, 0x61, 0x70,
		linv, linv, linv, linv,
		linv, linv, linv, linv,
	})
	upperBound := a.Data("decodeUpperBound", []byte{
		hinv, hinv, 0x2b, 0x39,
		0x4f, 0x5a, 0x6f, 0x7a,
		hinv, hinv, hinv, hinv,
		hinv, hinv, hinv, hinv,
	})
	shifts := a.Data("decodeShifts", []byte{
		0x00, 0x00, 0x3e, 0x34,
		0x00, 0x0f, 0x1a, 0x29,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	})
	nibble := a.Data("decodeNibble", repeat(0x0f, 16))
	x2f := a.Data("decode2f", repeat(0x2f, 16))
	var c3 byte = 3
	x2fOffset := a.Data("decode2fOffset", repeat(-c3, 16))
	merge := a.Data32("decodeMerge", []uint32{
		0x40014001,
		0x40014001,
		0x40014001,
		0x40014001,
		0x10000001,
		0x10000001,
		0x10000001,
		0x10000001,
	})

	a.NewFunction("decodeASM")
	a.NoSplit()

	dst := a.Argument("dst", 8)
	src := a.Argument("src", 8)
	length := a.Argument("len", 8)
	padding := a.Argument("padding", 4)
	url := a.Argument("url", 4)
	n := a.Argument("n", 8)
	ok := a.Argument("ok", 4)

	a.Start()

	bigloop_avx := a.NewLabel("bigloop_avx")
	loop := a.NewLabel("loop")
	ret := a.NewLabel("ret")
	invalid := a.NewLabel("invalid")

	d := &decode{
		a,

		asm.DI, asm.SI, asm.BX,

		ret, invalid,

		asm.X13, asm.X14,

		asm.X15,

		nibble,

		x2f, x2fOffset,

		asm.Address(asm.R15, 0), asm.Address(asm.R15, 16),
	}

	a.Movq(d.di, dst)
	a.Movq(d.si, src)
	a.Movq(d.cx, length)
	a.Movl(asm.DX, padding)
	a.Movb(asm.AX, url)

	a.Movq(asm.R14, d.si)
	a.Movq(asm.R15, d.di)

	a.Movw(asm.DX, asm.Constant(0xffff))

	a.Cmpq(asm.Constant(16+8), d.cx)
	a.Jb(loop)

	a.Movq(asm.R15, merge.Address())
	a.Movq(d.lowerBound, lowerBound)
	a.Movq(d.upperBound, upperBound)
	a.Movq(d.shifts, shifts)

	d.BigLoop(bigloop_avx)

	a.Label(loop)

	// loop

	a.Xorq(asm.AX, asm.AX)
	a.Jmp(invalid)

	a.Label(ret)

	a.Subq(d.di, asm.R15)

	a.Movq(n, d.di)
	a.Movb(ok, asm.Constant(1))

	a.Ret()

	a.Label(invalid)

	a.Bsfw(asm.AX, asm.AX)

	a.Subq(d.si, asm.R14)
	a.Addq(asm.AX, d.si)

	a.Movq(n, asm.AX)
	a.Movb(ok, asm.Constant(0))

	a.Ret()
}

func main() {
	if err := asm.Do("base64_encode_amd64.s", encodeHeader, encodeASM); err != nil {
		panic(err)
	}

	if err := asm.Do("base64_decode_amd64.s", decodeHeader, decodeASM); err != nil {
		panic(err)
	}
}
