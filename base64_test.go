// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package base64

import (
	ref "encoding/base64"
	"encoding/hex"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func testEncode(t *testing.T, enc Encoding, ref *ref.Encoding) {
	if err := quick.CheckEqual(ref.EncodeToString, enc.EncodeToString, &quick.Config{
		Values: func(args []reflect.Value, rand *rand.Rand) {
			off := rand.Intn(32)

			//data := make([]byte, 1 + rand.Intn(1024*1024) + off)
			data := make([]byte, 1+rand.Intn(32)+off)
			rand.Read(data[off:])
			args[0] = reflect.ValueOf(data[off:])
		},

		MaxCountScale: 1.75,
	}); err != nil {
		t.Error(err)
	}
}

func TestEncode(t *testing.T) {
	t.Run("Std", func(t *testing.T) {
		testEncode(t, StdEncoding, ref.StdEncoding)
	})

	t.Run("URL", func(t *testing.T) {
		testEncode(t, URLEncoding, ref.URLEncoding)
	})

	t.Run("RawStd", func(t *testing.T) {
		testEncode(t, RawStdEncoding, ref.RawStdEncoding)
	})

	t.Run("RawURL", func(t *testing.T) {
		testEncode(t, RawURLEncoding, ref.RawURLEncoding)
	})
}

func testDecode(t *testing.T, enc Encoding, scale float64, maxsize int) {
	if err := quick.CheckEqual(func(s string) (string, error) {
		b, err := ref.RawStdEncoding.DecodeString(s)
		return hex.EncodeToString(b), err
	}, func(s string) (string, error) {
		b, err := enc.DecodeString(s)
		return hex.EncodeToString(b), err
	}, &quick.Config{
		Values: func(args []reflect.Value, rand *rand.Rand) {
			//src := make([]byte, 1+rand.Intn(maxsize))
			src := make([]byte, maxsize)
			rand.Read(src)
			data := enc.EncodeToString(src)
			args[0] = reflect.ValueOf(data)
		},

		MaxCountScale: scale,
	}); err != nil {
		t.Error(err)
	}
}

func TestDecode(t *testing.T) {
	/*t.Run("Short", func(t *testing.T) {
		testDecode(t, StdEncoding, 1000, 7)
	})*/

	t.Run("RawStd", func(t *testing.T) {
		testDecode(t, RawStdEncoding, 2, 24)
	})

	/*t.Run("Std", func(t *testing.T) {
		testDecode(t, StdEncoding, 2, 1024*1024)
	})

	t.Run("URL", func(t *testing.T) {
		testDecode(t, URLEncoding, 2, 1024*1024)
	})*/
}

type size struct {
	name string
	l    int
}

var sizes = []size{
	{"32", 32},
	{"128", 128},
	{"1K", 1 * 1024},
	{"16K", 16 * 1024},
	{"128K", 128 * 1024},
	{"1M", 1024 * 1024},
	{"16M", 16 * 1024 * 1024},
	{"128M", 128 * 1024 * 1024},
}

func BenchmarkEncode(b *testing.B) {
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			src := make([]byte, size.l)
			rand.Read(src)

			dst := make([]byte, StdEncoding.EncodedLen(size.l))

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				StdEncoding.Encode(dst, src)
			}
		})
	}
}

func BenchmarkRefEncode(b *testing.B) {
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			src := make([]byte, size.l)
			rand.Read(src)

			dst := make([]byte, ref.StdEncoding.EncodedLen(size.l))

			b.SetBytes(int64(size.l))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ref.StdEncoding.Encode(dst, src)
			}
		})
	}
}
