package main

import (
	"fmt"
	"io"
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
