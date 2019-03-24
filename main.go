package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func badRequest(w http.ResponseWriter, err error) {
	log.Printf("bad request: %s\n", err)
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func internalError(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %s", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
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

			_, err = io.CopyN(w, f, length)
			if err != nil {
				internalError(w, fmt.Errorf("could not copy data: %s", err))
				return
			}

			err = f.Close()
			if err != nil {
				internalError(w, fmt.Errorf("could not close file: %s", err))
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
	w.Header().Set("Content-Type", "application/octet-stream")

	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	repo := parts[0]
	// TODO check if \w+

	store, err := NewFileStore(path.Join(os.Getenv("LOGSET_STORE"), repo))
	if err != nil {
		log.Printf("error creating store: %s\n", err)
		http.NotFound(w, r)
		return
	}

	switch endpoint := parts[1]; endpoint {
	case "offsets":
		handleOffsets(store, w, r)
	case "pull":
		handlePull(store, w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	if os.Getenv("LOGSET_STORE") == "" {
		log.Fatal("LOGSET_STORE is not set, exiting")
	}

	log.Println("server started")

	http.HandleFunc("/", route)
	log.Fatal(http.ListenAndServe(":4004", nil))
}
