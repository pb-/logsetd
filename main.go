package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
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
