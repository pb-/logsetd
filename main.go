package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

type contextKey string

// writeLock ensures there is only one writer; note that reading can happen concurrently
var writeLock sync.Mutex
var configKey = contextKey("config")
var alnum = regexp.MustCompile(`^[[:alnum:]]+$`)

func route(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value(configKey).(*config)
	w.Header().Set("Content-Type", "application/octet-stream")

	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	repo := parts[0]
	if !alnum.MatchString(repo) {
		http.NotFound(w, r)
		return
	}

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

type config struct {
	storePath string
}

func newConfig() *config {
	c := &config{
		storePath: os.Getenv("LOGSET_STORE"),
	}

	if c.storePath == "" {
		log.Fatal("LOGSET_STORE not set, exiting")
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
