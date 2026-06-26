package app

import (
	"context"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

type OrganizationService struct {
	store *store.Store
}

type CreateOrganizationOptions struct {
	ID   string
	Name string
}

func (s *OrganizationService) Create(ctx context.Context, opts CreateOrganizationOptions) (*domain.Organization, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}

	name := cleanText(opts.Name)
	if name == "" {
		return nil, invalidOptions("organization name is required")
	}

	id := domain.ID(cleanText(opts.ID))
	if id == "" {
		var err error
		id, err = newID("organization")
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	organization := domain.Organization{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateOrganization(organization); err != nil {
		return nil, err
	}
	return &organization, nil
}

func (s *OrganizationService) Get(ctx context.Context, id string) (*domain.Organization, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	organizationID, err := requiredID("organization id", id)
	if err != nil {
		return nil, err
	}
	return s.store.GetOrganization(organizationID)
}

func (s *OrganizationService) List(ctx context.Context) ([]domain.Organization, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	return s.store.ListOrganizations()
}
