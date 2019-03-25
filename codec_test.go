package main

import (
	"bufio"
	"io"
	"reflect"
	"strings"
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

func TestOffsetsBad(t *testing.T) {
	badInputs := []string{
		"name_ 0\n\n",
		"name -1\n\n",
	}

	for _, input := range badInputs {
		_, err := ReadOffsets(bufio.NewReader(strings.NewReader(input)))
		if err == nil {
			t.Fatalf("should fail: %s", input)
		}
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

func TestSliceInfoBad(t *testing.T) {
	badInputs := []string{
		"name_ 0 0\n",
		"name -1 0\n",
		"name 0 -1\n",
	}

	for _, input := range badInputs {
		_, _, _, err := ReadSliceInfo(bufio.NewReader(strings.NewReader(input)))
		if err == nil {
			t.Fatalf("should fail: %s", input)
		}
	}
}
