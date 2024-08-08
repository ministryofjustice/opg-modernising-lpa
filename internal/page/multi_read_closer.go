package page

// Modified from io/multi.go which carries the following licence:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"errors"
	"io"
)

type eofReadCloser struct{}

func (eofReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (eofReadCloser) Close() error {
	return nil
}

type multiReadCloser struct {
	readClosers []io.ReadCloser
}

func (mr *multiReadCloser) Read(p []byte) (n int, err error) {
	for len(mr.readClosers) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(mr.readClosers) == 1 {
			if r, ok := mr.readClosers[0].(*multiReadCloser); ok {
				mr.readClosers = r.readClosers
				continue
			}
		}
		n, err = mr.readClosers[0].Read(p)
		if err == io.EOF {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mr.readClosers[0] = eofReadCloser{} // permit earlier GC
			mr.readClosers = mr.readClosers[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mr.readClosers) > 0 {
				// Don't return EOF yet. More readClosers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

func (mr *multiReadCloser) WriteTo(w io.Writer) (sum int64, err error) {
	return mr.writeToWithBuffer(w, make([]byte, 1024*32))
}

func (mr *multiReadCloser) writeToWithBuffer(w io.Writer, buf []byte) (sum int64, err error) {
	for i, r := range mr.readClosers {
		var n int64
		if subMr, ok := r.(*multiReadCloser); ok { // reuse buffer with nested multiReaders
			n, err = subMr.writeToWithBuffer(w, buf)
		} else {
			n, err = io.CopyBuffer(w, r, buf)
		}
		sum += n
		if err != nil {
			mr.readClosers = mr.readClosers[i:] // permit resume / retry after error
			return sum, err
		}
		mr.readClosers[i] = nil // permit early GC
	}
	mr.readClosers = nil
	return sum, nil
}

func (mr *multiReadCloser) Close() error {
	var gerr error

	for _, rc := range mr.readClosers {
		if err := rc.Close(); err != nil {
			gerr = errors.Join(gerr, err)
		}
	}

	return gerr
}

var _ io.WriterTo = (*multiReadCloser)(nil)

// newMultiReadCloser returns a io.ReadCloser that's the logical concatenation of
// the provided input readClosers. They're read sequentially. Once all inputs
// have returned EOF, Read will return EOF.  If any of the readClosers return a
// non-nil, non-EOF error, Read will return that error.
func newMultiReadCloser(readClosers ...io.ReadCloser) io.ReadCloser {
	r := make([]io.ReadCloser, len(readClosers))
	copy(r, readClosers)
	return &multiReadCloser{r}
}
