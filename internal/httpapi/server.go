package httpapi

import (
	"net/http"

	"ghgo/internal/app"
)

type server struct {
	services *app.Services
}

func New(services *app.Services) http.Handler {
	s := &server{services: services}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /organizations", s.listOrganizations)
	mux.HandleFunc("POST /organizations", s.createOrganization)
	mux.HandleFunc("GET /organizations/{id}", s.getOrganization)
	mux.HandleFunc("GET /organizations/{organizationID}/facilities", s.listFacilities)
	mux.HandleFunc("POST /organizations/{organizationID}/facilities", s.createFacility)
	mux.HandleFunc("GET /organizations/{organizationID}/reporting-periods", s.listReportingPeriods)
	mux.HandleFunc("POST /organizations/{organizationID}/reporting-periods", s.createReportingPeriod)
	mux.HandleFunc("GET /reporting-periods/{id}", s.getReportingPeriod)
	mux.HandleFunc("GET /reporting-periods/{id}/settings", s.getReportingPeriodSettings)
	mux.HandleFunc("PUT /reporting-periods/{id}/settings", s.upsertReportingPeriodSettings)
	mux.HandleFunc("GET /reporting-periods/{id}/facilities/{facilityID}/electricity-settings", s.getElectricitySettings)
	mux.HandleFunc("PUT /reporting-periods/{id}/facilities/{facilityID}/electricity-settings", s.upsertElectricitySettings)
	mux.HandleFunc("POST /inputs/parse", s.parseInput)
	mux.HandleFunc("POST /inputs/commit", s.commitInput)
	mux.HandleFunc("POST /reporting-periods/{id}/calculations", s.runCalculation)
	mux.HandleFunc("GET /reporting-periods/{id}/calculation-runs", s.listCalculationRuns)
	mux.HandleFunc("GET /calculation-runs/{id}", s.getCalculationRun)
	mux.HandleFunc("GET /calculation-runs/{id}/report-tables", s.getReportTables)
	mux.HandleFunc("GET /factor-sets", s.listFactorSets)
	mux.HandleFunc("GET /factor-sets/default", s.getDefaultFactorSet)
	mux.HandleFunc("GET /factor-sets/{id}", s.getFactorSet)

	return withJSONFallback(mux)
}

func (s *server) healthz(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) listOrganizations(w http.ResponseWriter, r *http.Request) {
	organizations, err := s.services.Organizations.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newOrganizationResponses(organizations))
}

func (s *server) createOrganization(w http.ResponseWriter, r *http.Request) {
	var req createOrganizationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	organization, err := s.services.Organizations.Create(r.Context(), app.CreateOrganizationOptions{ID: req.ID, Name: req.Name})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, newOrganizationResponse(organization))
}

func (s *server) getOrganization(w http.ResponseWriter, r *http.Request) {
	organization, err := s.services.Organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newOrganizationResponse(organization))
}

func (s *server) listFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := s.services.Facilities.ListByOrganization(r.Context(), r.PathValue("organizationID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newFacilityResponses(facilities))
}

func (s *server) createFacility(w http.ResponseWriter, r *http.Request) {
	var req createFacilityRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	facility, err := s.services.Facilities.Create(r.Context(), app.CreateFacilityOptions{
		ID:             req.ID,
		OrganizationID: r.PathValue("organizationID"),
		Name:           req.Name,
		CountryCode:    req.CountryCode,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, newFacilityResponse(facility))
}

func (s *server) listReportingPeriods(w http.ResponseWriter, r *http.Request) {
	periods, err := s.services.ReportingPeriods.ListByOrganization(r.Context(), r.PathValue("organizationID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newReportingPeriodResponses(periods))
}

func (s *server) createReportingPeriod(w http.ResponseWriter, r *http.Request) {
	var req createReportingPeriodRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	startsOn, err := parseOptionalTime(req.StartsOn, "starts_on")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	endsOn, err := parseOptionalTime(req.EndsOn, "ends_on")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	period, err := s.services.ReportingPeriods.Create(r.Context(), app.CreateReportingPeriodOptions{
		ID:             req.ID,
		OrganizationID: r.PathValue("organizationID"),
		Year:           req.Year,
		StartsOn:       startsOn,
		EndsOn:         endsOn,
		Status:         req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, newReportingPeriodResponse(period))
}

func (s *server) getReportingPeriod(w http.ResponseWriter, r *http.Request) {
	period, err := s.services.ReportingPeriods.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newReportingPeriodResponse(period))
}

func (s *server) getReportingPeriodSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.services.ReportingPeriods.GetSettings(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newReportingPeriodSettingsResponse(settings))
}

func (s *server) upsertReportingPeriodSettings(w http.ResponseWriter, r *http.Request) {
	var req upsertReportingPeriodSettingsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	settings, err := s.services.ReportingPeriods.UpsertSettings(r.Context(), app.UpsertReportingPeriodSettingsOptions{
		ID:                req.ID,
		OrganizationID:    req.OrganizationID,
		ReportingPeriodID: r.PathValue("id"),
		MobileMethod:      req.MobileMethod,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newReportingPeriodSettingsResponse(settings))
}

func (s *server) getElectricitySettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.services.ReportingPeriods.GetElectricitySettings(r.Context(), r.PathValue("id"), r.PathValue("facilityID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newElectricitySettingsResponse(settings))
}

func (s *server) upsertElectricitySettings(w http.ResponseWriter, r *http.Request) {
	var req upsertElectricitySettingsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	cancelledAt, err := parseOptionalTimePtr(req.GOCancelledAt, "go_cancelled_at")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	settings, err := s.services.ReportingPeriods.UpsertElectricitySettings(r.Context(), app.UpsertElectricitySettingsOptions{
		ID:                    req.ID,
		OrganizationID:        req.OrganizationID,
		ReportingPeriodID:     r.PathValue("id"),
		FacilityID:            r.PathValue("facilityID"),
		HasGuaranteesOfOrigin: req.HasGuaranteesOfOrigin,
		GOCoverage:            req.GOCoverage,
		GOReference:           req.GOReference,
		GOMarket:              req.GOMarket,
		GOCancelledAt:         cancelledAt,
		EvidenceFileID:        req.EvidenceFileID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newElectricitySettingsResponse(settings))
}

func (s *server) parseInput(w http.ResponseWriter, r *http.Request) {
	var req parseInputRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.services.Inputs.Parse(r.Context(), app.ParseInputOptions{InputKind: req.InputKind, RawText: req.RawText})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newParseResultPayload(result))
}

func (s *server) commitInput(w http.ResponseWriter, r *http.Request) {
	var req commitInputRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	commitContext, err := req.Context.toInput()
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := s.services.Inputs.ParseAndCommit(r.Context(), app.ParseAndCommitInputOptions{
		Context: commitContext,
		RawText: req.RawText,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newCommitResultResponse(result))
}

func (s *server) runCalculation(w http.ResponseWriter, r *http.Request) {
	var req runCalculationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.services.Calculations.Run(r.Context(), app.RunCalculationOptions{
		OrganizationID:    req.OrganizationID,
		ReportingPeriodID: r.PathValue("id"),
		FactorSetID:       req.FactorSetID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, newCalculationResultResponse(result))
}

func (s *server) listCalculationRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.services.Calculations.ListRunsByPeriod(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newCalculationRunResponses(runs))
}

func (s *server) getCalculationRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.services.Calculations.GetRun(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newCalculationRunResponse(run))
}

func (s *server) getReportTables(w http.ResponseWriter, r *http.Request) {
	tables, err := s.services.Reports.BuildTables(r.Context(), app.BuildReportTablesOptions{CalculationRunID: r.PathValue("id")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newReportTablesResponse(tables))
}

func (s *server) listFactorSets(w http.ResponseWriter, r *http.Request) {
	factorSets, err := s.services.Factors.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newFactorSetResponses(factorSets))
}

func (s *server) getDefaultFactorSet(w http.ResponseWriter, r *http.Request) {
	factorSet, err := s.services.Factors.Default(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newFactorSetResponse(factorSet))
}

func (s *server) getFactorSet(w http.ResponseWriter, r *http.Request) {
	factorSet, err := s.services.Factors.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, newFactorSetResponse(factorSet))
}
