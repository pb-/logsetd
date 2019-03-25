package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type contextKey string

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
	port      int
}

func newConfig() *config {
	c := &config{
		storePath: os.Getenv("LOGSETD_STORE"),
	}

	if c.storePath == "" {
		log.Fatal("LOGSETD_STORE not set, exiting")
	}

	if port := os.Getenv("LOGSETD_PORT"); port != "" {
		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			log.Fatalf("LOGSETD_PORT malformed: %s, exiting", err)
		}
		c.port = int(p)
	} else {
		c.port = 4004
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

	log.Printf("server started on port %d", config.port)

	http.HandleFunc("/", withConfig(config, route))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", config.port), nil))
}
