package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type Store interface {
	Offsets() (map[string]int64, error)
	ReaderFrom(name string, offset int64) (io.ReadCloser, error)
	Appender(name string) (io.WriteCloser, error)
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
		if !file.IsDir() {
			offsets[file.Name()] = file.Size()
		}
	}

	return offsets, nil
}

func (f *fileStore) ReaderFrom(name string, offset int64) (io.ReadCloser, error) {
	filename := path.Join(f.path, name)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", filename, err)
	}

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek %s: %s", filename, err)
	}

	return file, nil
}

func (f *fileStore) Appender(name string) (io.WriteCloser, error) {
	filename := path.Join(f.path, name)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", filename, err)
	}

	return file, nil
}
