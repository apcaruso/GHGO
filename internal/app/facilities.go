package app

import (
	"context"
	"time"

	"ghgo/internal/domain"
	"ghgo/internal/ports"
)

type FacilityService struct {
	store ports.Store
}

type CreateFacilityOptions struct {
	ID             string
	OrganizationID string
	Name           string
	CountryCode    string
}

func (s *FacilityService) Create(ctx context.Context, opts CreateFacilityOptions) (*domain.Facility, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}

	organizationID, err := requiredID("organization id", opts.OrganizationID)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetOrganization(ctx, organizationID); err != nil {
		return nil, err
	}

	name := cleanText(opts.Name)
	if name == "" {
		return nil, invalidOptions("facility name is required")
	}
	countryCode, err := validateCountryCode(opts.CountryCode)
	if err != nil {
		return nil, err
	}

	id := domain.ID(cleanText(opts.ID))
	if id == "" {
		id, err = newID("facility")
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	facility := domain.Facility{
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		CountryCode:    countryCode,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.store.CreateFacility(ctx, facility); err != nil {
		return nil, err
	}
	return &facility, nil
}

func (s *FacilityService) ListByOrganization(ctx context.Context, organizationID string) ([]domain.Facility, error) {
	if err := checkStore(ctx, s.store); err != nil {
		return nil, err
	}
	id, err := requiredID("organization id", organizationID)
	if err != nil {
		return nil, err
	}
	return s.store.ListFacilitiesByOrganization(ctx, id)
}
