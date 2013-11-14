// FoundationDB Go Subspace Layer
// Copyright (c) 2013 FoundationDB, LLC

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// FIXME: document
package subspace

import (
	"github.com/FoundationDB/fdb-go/fdb"
	"github.com/FoundationDB/fdb-go/fdb/tuple"
	"bytes"
	"errors"
)

type Subspace interface {
	Sub(el ...interface{}) Subspace

	Bytes() []byte

	Pack(t tuple.Tuple) fdb.Key
	Unpack(k fdb.KeyConvertible) (tuple.Tuple, error)

	Contains(k fdb.KeyConvertible) bool

	fdb.KeyConvertible
	fdb.ExactRange
}

type subspace struct {
	b []byte
}

func AllKeys() Subspace {
	return subspace{}
}

func FromTuple(t tuple.Tuple) Subspace {
	return subspace{t.Pack()}
}

func FromBytes(b []byte) Subspace {
	s := make([]byte, len(b))
	copy(s, b)
	return subspace{b}
}

func (s subspace) Sub(el ...interface{}) Subspace {
	return subspace{concat(s.Bytes(), tuple.Tuple(el).Pack()...)}
}

func (s subspace) Bytes() []byte {
	return s.b
}

func (s subspace) Pack(t tuple.Tuple) fdb.Key {
	return fdb.Key(concat(s.b, t.Pack()...))
}

func (s subspace) Unpack(k fdb.KeyConvertible) (tuple.Tuple, error) {
	key := k.ToFDBKey()
	if !bytes.HasPrefix(key, s.b) {
		return nil, errors.New("key is not in subspace")
	}
	return tuple.Unpack(key[len(s.b):])
}

func (s subspace) Contains(k fdb.KeyConvertible) bool {
	return bytes.HasPrefix(k.ToFDBKey(), s.b)
}

func (s subspace) ToFDBKey() fdb.Key {
	return fdb.Key(s.b)
}

func (s subspace) BeginKey() fdb.Key {
	return concat(s.b, 0x00)
}

func (s subspace) EndKey() fdb.Key {
	return concat(s.b, 0xFF)
}

func (s subspace) BeginKeySelector() fdb.KeySelector {
	return fdb.FirstGreaterOrEqual(s.BeginKey())
}

func (s subspace) EndKeySelector() fdb.KeySelector {
	return fdb.FirstGreaterOrEqual(s.EndKey())
}

func concat(a []byte, b ...byte) []byte {
	r := make([]byte, len(a) + len(b))
	copy(r, a)
	copy(r[len(a):], b)
	return r
}