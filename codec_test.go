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
