package internal_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	. "bdd.fi/x/runitor/internal"

	"testing"
)

const RCap = 8

func TestWrite(t *testing.T) {
	rb := NewRingBuffer(RCap)
	if rbcap := rb.Cap(); rbcap != RCap {
		t.Errorf("expected Cap() to return %d but got %d", RCap, rbcap)
	}

	type test struct {
		name string
		str  string
		buf  string
	}

	tests := map[string]struct {
		str string
		buf string
	}{
		"simple write":    {str: "abc", buf: "abc"},
		"wrap":            {str: "012345", buf: "5bc01234"},
		"overrun discard": {str: "0123456789", buf: "78923456"},
		"zero byte write": {str: "", buf: "78923456"},
	}

	lenExp := 0

	for name, tc := range tests {
		n, err := fmt.Fprint(rb, tc.str)
		if err != nil {
			t.Errorf("%s: expected Write to succeed, got err '%v'", name, err)
		}

		if n != len(tc.str) {
			t.Errorf("%s: expected Write to return %d, got %d", name, len(tc.str), n)
		}

		lenExp = (lenExp + n)
		if lenExp > rb.Cap() {
			lenExp = rb.Cap()
		}

		if rblen := rb.Len(); rblen != lenExp {
			t.Errorf("%s: expected Len to return %d, got %d", name, lenExp, rblen)
		}

		snap := string(rb.Snapshot())
		if tc.buf != snap {
			t.Errorf("%s: expected ring buffer to be '%s', got '%s'", name, tc.buf, snap)
		}
	}
}

func TestNoWriteAfterRead(t *testing.T) {
	rb := NewRingBuffer(RCap)
	rb.Write([]byte{1})
	ioutil.ReadAll(rb)

	_, err := rb.Write([]byte{2})
	if err == nil || !errors.Is(err, ErrReadOnly) {
		t.Errorf("expected ring buffer to become read only after first read and receive ErrReadOnly but got err '%v'", err)
	}
}

func TestWriteAllocs(t *testing.T) {
	rb := NewRingBuffer(RCap)
	tb := make([]byte, RCap+1)
	allocs := testing.AllocsPerRun(1, func() {
		rb.Write(tb)
	})

	if allocs != 0 {
		t.Errorf("expected 0 allocations, observed %f\n", allocs)
	}
}

func TestReadAllocs(t *testing.T) {
	rb := NewRingBuffer(RCap)
	rb.Write(make([]byte, RCap+1))
	p := make([]byte, RCap)

	allocs := testing.AllocsPerRun(1, func() {
		rb.Read(p)
	})

	if allocs != 0 {
		t.Errorf("expected 0 allocations, observed %f\n", allocs)
	}
}