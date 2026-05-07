package app

import (
	"context"
	"errors"

	"ghgo/internal/domain"
	"ghgo/internal/factors"
	"ghgo/internal/ports"
)

const defaultFactorSetID = "factor_set_defra_2025"

type FactorService struct {
	store ports.Store
}

func (s *FactorService) EnsureDefault(ctx context.Context) (*domain.FactorSet, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	return factors.EnsureDefaultFactors(ctx, s.store)
}

func (s *FactorService) Get(ctx context.Context, id string) (*domain.FactorSet, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	factorSetID, err := requiredID("factor set id", id)
	if err != nil {
		return nil, err
	}
	return s.store.GetFactorSet(ctx, factorSetID)
}

func (s *FactorService) List(ctx context.Context) ([]domain.FactorSet, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	return s.store.ListFactorSets(ctx)
}

func (s *FactorService) Default(ctx context.Context) (*domain.FactorSet, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	factorSet, err := s.store.GetFactorSet(ctx, defaultFactorSetID)
	if err == nil {
		return factorSet, nil
	}
	if !errors.Is(err, ports.ErrNotFound) {
		return nil, err
	}
	return s.EnsureDefault(ctx)
}
