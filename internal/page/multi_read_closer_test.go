// Modified from io/multi.go which carries the following licence:
//
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package page

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"strings"
	"testing"
)

func TestMultiReadCloser(t *testing.T) {
	var mr io.Reader
	var buf []byte
	nread := 0
	withFooBar := func(tests func()) {
		r1 := io.NopCloser(strings.NewReader("foo "))
		r2 := io.NopCloser(strings.NewReader(""))
		r3 := io.NopCloser(strings.NewReader("bar"))
		mr = newMultiReadCloser(r1, r2, r3)
		buf = make([]byte, 20)
		tests()
	}
	expectRead := func(size int, expected string, eerr error) {
		nread++
		n, gerr := mr.Read(buf[0:size])
		if n != len(expected) {
			t.Errorf("#%d, expected %d bytes; got %d",
				nread, len(expected), n)
		}
		got := string(buf[0:n])
		if got != expected {
			t.Errorf("#%d, expected %q; got %q",
				nread, expected, got)
		}
		if gerr != eerr {
			t.Errorf("#%d, expected error %v; got %v",
				nread, eerr, gerr)
		}
		buf = buf[n:]
	}
	withFooBar(func() {
		expectRead(2, "fo", nil)
		expectRead(5, "o ", nil)
		expectRead(5, "bar", nil)
		expectRead(5, "", io.EOF)
	})
	withFooBar(func() {
		expectRead(4, "foo ", nil)
		expectRead(1, "b", nil)
		expectRead(3, "ar", nil)
		expectRead(1, "", io.EOF)
	})
	withFooBar(func() {
		expectRead(5, "foo ", nil)
	})
}

func TestMultiReadCloserAsWriterTo(t *testing.T) {
	mr := newMultiReadCloser(
		io.NopCloser(strings.NewReader("foo ")),
		newMultiReadCloser( // Tickle the buffer reusing codepath
			io.NopCloser(strings.NewReader("")),
			io.NopCloser(strings.NewReader("bar")),
		),
	)
	mrAsWriterTo, ok := mr.(io.WriterTo)
	if !ok {
		t.Fatalf("expected cast to WriterTo to succeed")
	}
	sink := &strings.Builder{}
	n, err := mrAsWriterTo.WriteTo(sink)
	if err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
	if n != 7 {
		t.Errorf("expected read 7 bytes; got %d", n)
	}
	if result := sink.String(); result != "foo bar" {
		t.Errorf(`expected "foo bar"; got %q`, result)
	}
}

// readerFunc is a Reader implemented by the underlying func.
type readerFunc func(p []byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) {
	return f(p)
}

func (readerFunc) Close() error {
	return nil
}

// callDepth returns the logical call depth for the given PCs.
func callDepth(callers []uintptr) (depth int) {
	frames := runtime.CallersFrames(callers)
	more := true
	for more {
		_, more = frames.Next()
		depth++
	}
	return
}

// Test that MultiReadCloser properly flattens chained multiReaders when Read is called
func TestMultiReadCloserFlatten(t *testing.T) {
	pc := make([]uintptr, 1000) // 1000 should fit the full stack
	n := runtime.Callers(0, pc)
	var myDepth = callDepth(pc[:n])
	var readDepth int // will contain the depth from which fakeReader.Read was called
	var r io.ReadCloser = newMultiReadCloser(readerFunc(func(p []byte) (int, error) {
		n := runtime.Callers(1, pc)
		readDepth = callDepth(pc[:n])
		return 0, errors.New("irrelevant")
	}))

	// chain a bunch of multiReaders
	for i := 0; i < 100; i++ {
		r = newMultiReadCloser(r)
	}

	r.Read(nil) // don't care about errors, just want to check the call-depth for Read

	if readDepth != myDepth+2 { // 2 should be multiReader.Read and fakeReader.Read
		t.Errorf("multiReader did not flatten chained multiReaders: expected readDepth %d, got %d",
			myDepth+2, readDepth)
	}
}

// byteAndEOFReader is a Reader which reads one byte (the underlying
// byte) and EOF at once in its Read call.
type byteAndEOFReader byte

func (b byteAndEOFReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		// Read(0 bytes) is useless. We expect no such useless
		// calls in this test.
		panic("unexpected call")
	}
	p[0] = byte(b)
	return 1, io.EOF
}

// This used to yield bytes forever; issue 16795.
func TestMultiReadCloserSingleByteWithEOF(t *testing.T) {
	got, err := io.ReadAll(io.LimitReader(newMultiReadCloser(io.NopCloser(byteAndEOFReader('a')), io.NopCloser(byteAndEOFReader('b'))), 10))
	if err != nil {
		t.Fatal(err)
	}
	const want = "ab"
	if string(got) != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// Test that a reader returning (n, EOF) at the end of a MultiReadCloser
// chain continues to return EOF on its final read, rather than
// yielding a (0, EOF).
func TestMultiReadCloserFinalEOF(t *testing.T) {
	r := newMultiReadCloser(io.NopCloser(bytes.NewReader(nil)), io.NopCloser(byteAndEOFReader('a')))
	buf := make([]byte, 2)
	n, err := r.Read(buf)
	if n != 1 || err != io.EOF {
		t.Errorf("got %v, %v; want 1, EOF", n, err)
	}
}

func TestInterleavedMultiReadCloser(t *testing.T) {
	r1 := io.NopCloser(strings.NewReader("123"))
	r2 := io.NopCloser(strings.NewReader("45678"))

	mr1 := newMultiReadCloser(r1, r2)
	mr2 := newMultiReadCloser(mr1)

	buf := make([]byte, 4)

	// Have mr2 use mr1's []Readers.
	// Consume r1 (and clear it for GC to handle) and consume part of r2.
	n, err := io.ReadFull(mr2, buf)
	if got := string(buf[:n]); got != "1234" || err != nil {
		t.Errorf(`ReadFull(mr2) = (%q, %v), want ("1234", nil)`, got, err)
	}

	// Consume the rest of r2 via mr1.
	// This should not panic even though mr2 cleared r1.
	n, err = io.ReadFull(mr1, buf)
	if got := string(buf[:n]); got != "5678" || err != nil {
		t.Errorf(`ReadFull(mr1) = (%q, %v), want ("5678", nil)`, got, err)
	}
}

type rc struct {
	err    error
	closed bool
}

func (*rc) Read(p []byte) (int, error) {
	return 0, nil
}

func (r *rc) Close() error {
	r.closed = true
	return r.err
}

func TestMultiReadCloserClose(t *testing.T) {
	rc1 := &rc{}
	rc2 := &rc{}

	mr := newMultiReadCloser(rc1, rc2)
	if err := mr.Close(); err != nil {
		t.Errorf(`Close() = %v, want nil`, err)
	}
	if !rc1.closed {
		t.Error("rc1 not closed")
	}
	if !rc2.closed {
		t.Error("rc1 not closed")
	}
}

func TestMultiReadCloserCloseWhenErrors(t *testing.T) {
	err1 := errors.New("1")
	err2 := errors.New("2")
	expected := errors.Join(err1, err2)

	rc1 := &rc{err: err1}
	rc2 := &rc{err: err2}

	mr := newMultiReadCloser(rc1, rc2)
	if err := mr.Close(); err.Error() != expected.Error() {
		t.Errorf(`Close() = %v, want %v`, err, expected)
	}
	if !rc1.closed {
		t.Error("rc1 not closed")
	}
	if !rc2.closed {
		t.Error("rc1 not closed")
	}
}
