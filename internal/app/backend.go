package app

import (
	"context"
	"database/sql"
	"fmt"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/store"
)

type Backend struct {
	DB               *sql.DB
	Store            *store.Store
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
	defaultFactorSet, err := factors.EnsureDefaultFactors(ctx, st)
	if err != nil {
		return nil, fmt.Errorf("seed default factors: %w", err)
	}

	services, err := NewServices(st)
	if err != nil {
		return nil, err
	}

	closeOnError = false
	return &Backend{
		DB:               db,
		Store:            st,
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
