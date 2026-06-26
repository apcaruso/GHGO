package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ghgo/internal/app"
	"ghgo/internal/httpapi"
)

const (
	defaultDBPath   = "data/ghgo.sqlite"
	defaultHTTPAddr = "127.0.0.1:8080"
)

func main() {
	dbPath := os.Getenv("GHGO_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	addr := os.Getenv("GHGO_HTTP_ADDR")
	if addr == "" {
		addr = defaultHTTPAddr
	}

	backend, err := app.OpenSQLite(context.Background(), dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghgo-api: initialize backend: %v\n", err)
		os.Exit(1)
	}
	defer backend.Close()

	server := &http.Server{
		Addr:              addr,
		Handler:           newHandler(httpapi.New(backend.Services)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Fprintf(os.Stdout, "ghgo API listening\naddress: %s\ndatabase: %s\ndefault factor set: %s\n", addr, dbPath, backend.DefaultFactorSet.ID)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "ghgo-api: serve: %v\n", err)
		os.Exit(1)
	}
}

func newHandler(api http.Handler) http.Handler {
	uiDir := uiDirectory()
	if uiDir == "" {
		return api
	}

	mux := http.NewServeMux()
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(uiDir))))
	mux.HandleFunc("GET /ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
	})
	mux.Handle("/", api)
	return mux
}

func uiDirectory() string {
	if dir := usableDirectory(os.Getenv("GHGO_UI_DIR")); dir != "" {
		return dir
	}

	for _, candidate := range []string{"../frontend", "frontend"} {
		if dir := usableDirectory(candidate); dir != "" {
			return dir
		}
	}
	return ""
}

func usableDirectory(path string) string {
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
