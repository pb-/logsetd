package main

import (
	"bufio"
	"errors"
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

		name := parts[0]
		offset, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, err
		}

		if offset < 0 {
			return nil, errors.New("offset must be positive")
		}

		if !alnum.MatchString(name) {
			return nil, errors.New("bad name")
		}

		offsets[name] = offset
	}
}

func WriteSliceInfo(w io.Writer, name string, offset int64, length int64) error {
	_, err := fmt.Fprintf(w, "%s %d %d\n", name, offset, length)
	if err != nil {
		return err
	}

	return nil
}

func ReadSliceInfo(r *bufio.Reader) (name string, offset int64, length int64, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) != 3 {
		err = fmt.Errorf("bad slice info: %d parts", len(parts))
		return
	}

	name = parts[0]
	if !alnum.MatchString(name) {
		err = errors.New("bad name")
		return
	}

	offset, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}
	if offset < 0 {
		err = errors.New("offset must be positive")
		return
	}

	length, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return
	}
	if length < 0 {
		err = errors.New("length must be positive")
		return
	}

	return
}
