// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 gccgo appengine

package base64

import (
	ref "encoding/base64"
	"io"
)

type Encoding struct {
	impl *ref.Encoding
}

func newEncoding(encType encodingType) Encoding {
	switch encType {
	case encodeStd:
		return Encoding{ref.StdEncoding}
	case encodeURL:
		return Encoding{ref.URLEncoding}
	default:
		panic("invalid encoding type")
	}
}

func (enc Encoding) WithPadding(padding rune) Encoding {
	return Encoding{enc.impl.WithPadding(padding)}
}

func (enc Encoding) Decode(dst, src []byte) (n int, err error) {
	n, err = enc.impl.Decode(dst, src)
	if _, ok := err.(ref.CorruptInputError); ok {
		err = ErrFormat
	}

	return
}

func (enc Encoding) DecodeString(s string) (b []byte, err error) {
	b, err = enc.impl.DecodeString(s)
	if _, ok := err.(ref.CorruptInputError); ok {
		err = ErrFormat
	}

	return
}

func (enc Encoding) DecodedLen(n int) int {
	return enc.impl.DecodedLen(n)
}

func (enc Encoding) Encode(dst, src []byte) {
	return enc.impl.Encode(dst, src)
}

func (enc Encoding) EncodeToString(src []byte) string {
	return enc.impl.EncodeToString(src)
}

func (enc Encoding) EncodedLen(n int) int {
	return enc.impl.EncodedLen(n)
}

func NewDecoder(enc Encoding, r io.Reader) io.Reader {
	return ref.NewDecoder(enc.impl, r)
}

func NewEncoder(enc Encoding, w io.Writer) io.WriteCloser {
	return ref.NewEncoder(enc.impl, w)
}
