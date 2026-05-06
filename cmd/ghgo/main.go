package main

import (
	"context"
	"fmt"
	"os"

	"ghgo/internal/factors"
	"ghgo/internal/store"
	"ghgo/internal/ui"
)

const defaultDBPath = "data/ghgo.sqlite"

func main() {
	dbPath := os.Getenv("GHGO_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghgo: open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := store.RunMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "ghgo: run migrations: %v\n", err)
		os.Exit(1)
	}

	st := store.New(db)
	if _, err := factors.EnsureDefaultFactors(context.Background(), st); err != nil {
		fmt.Fprintf(os.Stderr, "ghgo: seed default factors: %v\n", err)
		os.Exit(1)
	}

	if err := ui.Run(dbPath, st); err != nil {
		fmt.Fprintf(os.Stderr, "ghgo: start UI: %v\n", err)
		os.Exit(1)
	}
}
