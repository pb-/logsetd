package main

import (
	//"fmt"
	//"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func handleOffsets(store Store, w http.ResponseWriter, r *http.Request) {
	offsets, err := store.Offsets()
	if err != nil {
		log.Printf("could not get offsets: %s\n", err)
		http.NotFound(w, r)
		return
	}

	err = WriteOffsets(w, offsets)
	if err != nil {
		log.Printf("could write offsets: %s\n", err)
		http.NotFound(w, r)
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
	default:
		http.NotFound(w, r)
	}

	// fmt.Printf("%s\n", r.URL.Path)
	// io.Copy(w, r.Body)
}

func main() {
	if os.Getenv("LOGSET_STORE") == "" {
		log.Fatal("LOGSET_STORE is not set, exiting")
	}

	log.Println("server started")

	http.HandleFunc("/", route)
	log.Fatal(http.ListenAndServe(":4004", nil))
}
