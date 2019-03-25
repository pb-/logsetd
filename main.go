package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type contextKey string

// writeLock ensures there is only one writer; note that reading can happen concurrently
var writeLock sync.Mutex

var configKey = contextKey("config")

func badRequest(w http.ResponseWriter, err error) {
	log.Printf("bad request: %s", err)
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func internalError(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %s", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func forbidden(w http.ResponseWriter, err error) {
	log.Printf("forbidden: %s", err)
	http.Error(w, "Forbidden", http.StatusForbidden)
}

func handlePush(store Store, w http.ResponseWriter, req *http.Request) {
	writeLock.Lock()
	defer writeLock.Unlock()

	localOffsets, err := store.Offsets()
	if err != nil {
		internalError(w, fmt.Errorf("could not read local offsets: %s", err))
		return
	}

	r := bufio.NewReader(req.Body)
	for {
		name, offset, length, err := ReadSliceInfo(r)
		if err != nil {
			if err != io.EOF {
				badRequest(w, fmt.Errorf("error reading slice info: %s", err))
			}
			return
		}

		if offset != localOffsets[name] {
			forbidden(w, fmt.Errorf("offset mismatch: %d %d", offset, localOffsets[name]))
			return
		}

		f, err := store.Appender(name)
		if err != nil {
			internalError(w, fmt.Errorf("could not get appender: %s", err))
			return
		}
		defer f.Close()

		_, err = io.CopyN(f, r, length)
		if err != nil {
			internalError(w, fmt.Errorf("could not copy data: %s", err))
			return
		}
	}
}

func handlePull(store Store, w http.ResponseWriter, req *http.Request) {
	remoteOffsets, err := ReadOffsets(bufio.NewReader(req.Body))
	if err != nil {
		badRequest(w, err)
		return
	}

	localOffsets, err := store.Offsets()
	if err != nil {
		internalError(w, fmt.Errorf("could not read local offsets: %s", err))
		return
	}

	err = WriteOffsets(w, localOffsets)
	if err != nil {
		internalError(w, fmt.Errorf("could not write local offsets: %s", err))
		return
	}

	for name, offset := range localOffsets {
		if remoteOffsets[name] < offset {
			remoteOffset := remoteOffsets[name]
			length := offset - remoteOffset
			err = WriteSliceInfo(w, name, remoteOffset, length)
			if err != nil {
				internalError(w, fmt.Errorf("could not write slice info: %s", err))
				return
			}

			f, err := store.ReaderFrom(name, remoteOffset)
			if err != nil {
				internalError(w, fmt.Errorf("could not get reader: %s", err))
				return
			}
			defer f.Close()

			_, err = io.CopyN(w, f, length)
			if err != nil {
				internalError(w, fmt.Errorf("could not copy data: %s", err))
				return
			}
		}
	}
}

func handleOffsets(store Store, w http.ResponseWriter, r *http.Request) {
	offsets, err := store.Offsets()
	if err != nil {
		internalError(w, fmt.Errorf("could not get offsets: %s", err))
		return
	}

	err = WriteOffsets(w, offsets)
	if err != nil {
		internalError(w, fmt.Errorf("could write offsets: %s", err))
		return
	}
}

func route(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value(configKey).(*config)
	w.Header().Set("Content-Type", "application/octet-stream")

	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	repo := parts[0]
	// TODO check if \w+

	store, err := NewFileStore(path.Join(c.storePath, repo))
	if err != nil {
		log.Printf("error creating store: %s", err)
		http.NotFound(w, r)
		return
	}

	switch endpoint := parts[1]; endpoint {
	case "offsets":
		handleOffsets(store, w, r)
	case "pull":
		handlePull(store, w, r)
	case "push":
		handlePush(store, w, r)
	default:
		http.NotFound(w, r)
	}
}

func randomKey(length int) string {
	key := make([]byte, length)
	rand.Read(key)
	return base32.StdEncoding.EncodeToString(key)
}

type config struct {
	storePath string
	initKey   string
}

func newConfig() *config {
	c := &config{
		storePath: os.Getenv("LOGSET_STORE"),
		initKey:   os.Getenv("LOGSET_INIT_KEY"),
	}

	if c.storePath == "" {
		log.Fatal("LOGSET_STORE not set, exiting")
	}

	if c.initKey == "" {
		c.initKey = randomKey(15)
		log.Printf("LOGSET_INIT_KEY not set, using random key %s", c.initKey)
	}

	return c
}

func withConfig(c *config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r.WithContext(context.WithValue(r.Context(), configKey, c)))
	}
}

func main() {
	config := newConfig()

	log.Println("server started")

	http.HandleFunc("/", withConfig(config, route))
	log.Fatal(http.ListenAndServe(":4004", nil))
}
