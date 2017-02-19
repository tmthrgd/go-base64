// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

package base64

import "io"

import "fmt"
import "strings"

type Encoding struct {
	url     bool
	padding rune
}

func newEncoding(encType encodingType) Encoding {
	switch encType {
	case encodeStd:
		return Encoding{false, StdPadding}
	case encodeURL:
		return Encoding{true, StdPadding}
	default:
		panic("invalid encoding type")
	}
}

func (enc Encoding) WithPadding(padding rune) Encoding {
	return Encoding{enc.url, padding}
}

// TODO: func (enc Encoding) Strict() Encoding {}

func (enc Encoding) EncodeToString(src []byte) string {
	buf := make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(buf, src)
	return string(buf)
}

func (enc Encoding) EncodedLen(n int) int {
	if enc.padding == NoPadding {
		return (n*8 + 5) / 6 // minimum # chars at 6 bits per char
	}

	return (n + 2) / 3 * 4 // minimum # 4-char quanta, 3 bytes each
}

func (enc Encoding) DecodeString(s string) ([]byte, error) {
	dbuf := make([]byte, enc.DecodedLen(len(s)))
	n, err := enc.Decode(dbuf, []byte(s))
	return dbuf[:n], err
}

func (enc Encoding) DecodedLen(n int) int {
	if enc.padding == NoPadding {
		// Unpadded data may end with partial block of 2-3 characters.
		return (n*6 + 7) / 8
	}

	// Padded base64 should always be a multiple of 4 characters in length.
	return n / 4 * 3
}

func NewDecoder(enc Encoding, r io.Reader) io.Reader {
	panic("not implemented")
}

func NewEncoder(enc Encoding, w io.Writer) io.WriteCloser {
	panic("not implemented")
}

func (enc Encoding) Encode(dst, src []byte) {
	if len(src) == 0 {
		return
	}

	encodeASM(&dst[0], &src[0], uint64(len(src)), enc.padding, enc.url)
}

func (enc Encoding) Decode(dst, src []byte) (n int, err error) {
	if len(src) == 0 {
		return
	}

	nn, ok := decodeASM(&dst[0], &src[0], uint64(len(dst)), enc.padding, enc.url)
	if !ok {
		err = ErrFormat
	}

	fmt.Println(" " + strings.Join(strings.Split(string(src), ""), " "))
	fmt.Printf("%x\n", dst)

	fmt.Println(nn, ok)

	n = int(nn)
	return
}

//go:generate go run asm_gen.go

// This function is implemented in base64_encode_amd64.s
//go:noescape
func encodeASM(dst *byte, src *byte, len uint64, padding int32, url bool)

// This function is implemented in base64_decode_amd64.s
//go:noescape
func decodeASM(dst *byte, src *byte, len uint64, padding int32, url bool) (n uint64, ok bool)
