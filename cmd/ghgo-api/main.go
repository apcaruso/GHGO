package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
		Handler:           httpapi.New(backend.Services),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Fprintf(os.Stdout, "ghgo API listening\naddress: %s\ndatabase: %s\ndefault factor set: %s\n", addr, dbPath, backend.DefaultFactorSet.ID)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "ghgo-api: serve: %v\n", err)
		os.Exit(1)
	}
}
