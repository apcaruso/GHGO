package app

import (
	"context"
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/ports"
	"ghgo/internal/store"
)

type Backend struct {
	DB               *sql.DB
	Store            ports.Store
	Services         *Services
	DefaultFactorSet *domain.FactorSet
}

func OpenSQLite(ctx context.Context, dbPath string) (*Backend, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	closeOnError := true
	defer func() {
		if closeOnError {
			db.Close()
		}
	}()

	if err := store.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	st := store.New(db)
	repo := store.NewRepository(st)
	defaultFactorSet, err := factors.EnsureDefaultFactors(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("seed default factors: %w", err)
	}

	services, err := NewServices(repo)
	if err != nil {
		return nil, err
	}

	closeOnError = false
	return &Backend{
		DB:               db,
		Store:            repo,
		Services:         services,
		DefaultFactorSet: defaultFactorSet,
	}, nil
}

func (b *Backend) Close() error {
	if b == nil || b.DB == nil {
		return nil
	}
	return b.DB.Close()
}
