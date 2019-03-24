package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Store interface {
	Offsets() (map[string]int64, error)
	ReaderFrom(name string, offset int64) (io.Reader, error)
	Appender(name string) (io.Writer, error)
}

type fileStore struct {
	path string
}

func NewFileStore(path string) (Store, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%s: not a valid store", path)
	}

	return &fileStore{
		path: path,
	}, nil
}

func (f *fileStore) Offsets() (map[string]int64, error) {
	files, err := ioutil.ReadDir(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", f.path, err)
	}

	offsets := map[string]int64{}
	for _, file := range files {
		offsets[file.Name()] = file.Size()
	}

	return offsets, nil
}

func (*fileStore) ReaderFrom(name string, offset int64) (io.Reader, error) {
	return nil, nil
}

func (*fileStore) Appender(name string) (io.Writer, error) {
	return nil, nil
}
