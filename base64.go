// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

// Package base64 is an efficient base64 implementation for Golang.
package base64

import "errors"

type encodingType int

const (
	encodeStd encodingType = iota
	encodeURL
)

const (
	StdPadding rune = '=' // Standard padding character
	NoPadding  rune = -1  // No padding
)

var (
	StdEncoding = newEncoding(encodeStd)
	URLEncoding = newEncoding(encodeURL)

	RawStdEncoding = StdEncoding.WithPadding(NoPadding)
	RawURLEncoding = URLEncoding.WithPadding(NoPadding)
)

var ErrFormat = errors.New("go-base64: invalid input")
