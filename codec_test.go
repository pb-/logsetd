package main

import (
	"bufio"
	"io"
	"reflect"
	"testing"
)

func TestOffsets(t *testing.T) {
	offsets := map[string]int64{
		"abc": 10,
		"def": 49581,
	}
	r, w := io.Pipe()

	go func() {
		WriteOffsets(w, offsets)
	}()

	read, err := ReadOffsets(bufio.NewReader(r))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(offsets, read) {
		t.Fatal("not equal")
	}
}

func TestSliceInfo(t *testing.T) {
	name := "foo"
	offset := int64(1)
	length := int64(9001)

	r, w := io.Pipe()
	go func() {
		WriteSliceInfo(w, name, offset, length)
	}()

	rName, rOffset, rLength, err := ReadSliceInfo(bufio.NewReader(r))
	if err != nil {
		t.Fatal(err)
	}

	if name != rName || offset != rOffset || length != rLength {
		t.Fatalf("not equal: %s %d %d", rName, rOffset, rLength)
	}
}
