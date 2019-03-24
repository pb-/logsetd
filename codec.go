package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func WriteOffsets(w io.Writer, offsets map[string]int64) error {
	for name, offset := range offsets {
		_, err := fmt.Fprintf(w, "%s %d\n", name, offset)
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w)
	if err != nil {
		return err
	}

	return nil
}

func ReadOffsets(r *bufio.Reader) (map[string]int64, error) {
	offsets := map[string]int64{}

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "\n" {
			return offsets, nil
		}

		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad offset line: %d parts", len(parts))
		}

		offset, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, err
		}

		offsets[parts[0]] = offset
	}
}

func WriteSliceInfo(w io.Writer, name string, offset int64, length int64) error {
	_, err := fmt.Fprintf(w, "%s %d %d\n", name, offset, length)
	if err != nil {
		return err
	}

	return nil
}
