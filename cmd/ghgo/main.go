package main

import (
	"context"
	"fmt"
	"os"

	"ghgo/internal/app"
)

const defaultDBPath = "data/ghgo.sqlite"

func main() {
	dbPath := os.Getenv("GHGO_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	backend, err := app.OpenSQLite(context.Background(), dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghgo: initialize backend: %v\n", err)
		os.Exit(1)
	}
	defer backend.Close()

	fmt.Fprintf(os.Stdout, "ghgo backend initialized\ndatabase: %s\ndefault factor set: %s\n", dbPath, backend.DefaultFactorSet.ID)
}
