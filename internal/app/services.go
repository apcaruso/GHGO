package app

import "ghgo/internal/store"

type Services struct {
	Organizations    *OrganizationService
	Facilities       *FacilityService
	ReportingPeriods *ReportingPeriodService
	Inputs           *InputService
	Calculations     *CalculationService
	Reports          *ReportService
	Factors          *FactorService
}

func NewServices(st *store.Store) (*Services, error) {
	if st == nil {
		return nil, invalidOptions("store is required")
	}
	return &Services{
		Organizations:    &OrganizationService{store: st},
		Facilities:       &FacilityService{store: st},
		ReportingPeriods: &ReportingPeriodService{store: st},
		Inputs:           &InputService{store: st},
		Calculations:     &CalculationService{store: st},
		Reports:          &ReportService{store: st},
		Factors:          &FactorService{store: st},
	}, nil
}
