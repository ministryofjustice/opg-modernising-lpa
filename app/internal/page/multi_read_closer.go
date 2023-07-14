// Modified from io/multi.go which carries the following licence:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package page

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
			n, err = copyBuffer(w, r, buf)
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

// MultiReadCloser returns a io.ReadCloser that's the logical concatenation of
// the provided input readClosers. They're read sequentially. Once all inputs
// have returned EOF, Read will return EOF.  If any of the readClosers return a
// non-nil, non-EOF error, Read will return that error.
func MultiReadCloser(readClosers ...io.ReadCloser) io.ReadCloser {
	r := make([]io.ReadCloser, len(readClosers))
	copy(r, readClosers)
	return &multiReadCloser{r}
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
