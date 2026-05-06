package ui

import (
	"context"
	"errors"

	"ghgo/internal/domain"
	"ghgo/internal/store"
)

type State struct {
	Store  *store.Store
	DBPath string

	Organization  *domain.Organization
	Organizations []domain.Organization

	Facilities []domain.Facility

	ReportingPeriods []domain.ReportingPeriod
	ReportingPeriod  *domain.ReportingPeriod
	Settings         *domain.ReportingPeriodSettings

	FactorSet  *domain.FactorSet
	FactorSets []domain.FactorSet

	CalculationRuns []domain.CalculationRun
}

func NewState(st *store.Store, dbPath string) *State {
	return &State{Store: st, DBPath: dbPath}
}

func (s *State) Refresh(ctx context.Context) error {
	if s == nil || s.Store == nil {
		return required("store")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	selectedOrganizationID := ""
	if s.Organization != nil {
		selectedOrganizationID = s.Organization.ID
	}
	selectedReportingPeriodID := ""
	if s.ReportingPeriod != nil {
		selectedReportingPeriodID = s.ReportingPeriod.ID
	}
	selectedFactorSetID := ""
	if s.FactorSet != nil {
		selectedFactorSetID = s.FactorSet.ID
	}

	organizations, err := s.Store.ListOrganizations()
	if err != nil {
		return err
	}
	s.Organizations = organizations
	s.Organization = nil
	for i := range organizations {
		if organizations[i].ID == selectedOrganizationID {
			s.Organization = &organizations[i]
			break
		}
	}
	if s.Organization == nil && len(organizations) > 0 {
		s.Organization = &organizations[0]
	}

	s.Facilities = nil
	s.ReportingPeriods = nil
	s.ReportingPeriod = nil
	s.Settings = nil
	s.CalculationRuns = nil
	if s.Organization != nil {
		facilities, err := s.Store.ListFacilitiesByOrganization(s.Organization.ID)
		if err != nil {
			return err
		}
		s.Facilities = facilities

		periods, err := s.Store.ListReportingPeriodsByOrganization(s.Organization.ID)
		if err != nil {
			return err
		}
		s.ReportingPeriods = periods
		for i := range periods {
			if periods[i].ID == selectedReportingPeriodID {
				s.ReportingPeriod = &periods[i]
				break
			}
		}
		if s.ReportingPeriod == nil && len(periods) > 0 {
			s.ReportingPeriod = &periods[len(periods)-1]
		}
		if s.ReportingPeriod != nil {
			settings, err := s.Store.GetReportingPeriodSettings(s.ReportingPeriod.ID)
			if errors.Is(err, store.ErrNotFound) {
				settings = nil
			} else if err != nil {
				return err
			}
			s.Settings = settings

			runs, err := s.Store.ListCalculationRunsByPeriod(s.ReportingPeriod.ID)
			if err != nil {
				return err
			}
			s.CalculationRuns = runs
		}
	}

	factorSets, err := s.Store.ListFactorSets()
	if err != nil {
		return err
	}
	s.FactorSets = factorSets
	s.FactorSet = nil
	for i := range factorSets {
		if factorSets[i].ID == selectedFactorSetID {
			s.FactorSet = &factorSets[i]
			break
		}
	}
	if s.FactorSet == nil && len(factorSets) > 0 {
		s.FactorSet = &factorSets[len(factorSets)-1]
	}

	return nil
}

func (s *State) SetOrganization(id string) {
	s.Organization = nil
	for i := range s.Organizations {
		if s.Organizations[i].ID == id {
			s.Organization = &s.Organizations[i]
			return
		}
	}
}

func (s *State) SetReportingPeriod(id string) {
	s.ReportingPeriod = nil
	for i := range s.ReportingPeriods {
		if s.ReportingPeriods[i].ID == id {
			s.ReportingPeriod = &s.ReportingPeriods[i]
			return
		}
	}
}

func (s *State) SetFactorSet(id string) {
	s.FactorSet = nil
	for i := range s.FactorSets {
		if s.FactorSets[i].ID == id {
			s.FactorSet = &s.FactorSets[i]
			return
		}
	}
}

func (s *State) CurrentReportingYear() int {
	if s.ReportingPeriod == nil {
		return 0
	}
	return s.ReportingPeriod.Year
}

func (s *State) LatestCompletedRun() *domain.CalculationRun {
	for i := len(s.CalculationRuns) - 1; i >= 0; i-- {
		if s.CalculationRuns[i].Status == domain.CalculationRunStatusCompleted {
			return &s.CalculationRuns[i]
		}
	}
	return nil
}

func (s *State) DefaultFactorSet() *domain.FactorSet {
	for i := range s.FactorSets {
		if s.FactorSets[i].Name == "DEFRA/DESNZ 2025" || (s.FactorSets[i].Source == "DEFRA" && s.FactorSets[i].Year == 2025 && s.FactorSets[i].Version == "2025") {
			return &s.FactorSets[i]
		}
	}
	return s.FactorSet
}
