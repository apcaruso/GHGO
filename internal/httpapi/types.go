package httpapi

import (
	"fmt"
	"strings"
	"time"

	"ghgo/internal/calc"
	"ghgo/internal/domain"
	"ghgo/internal/input"
	"ghgo/internal/report"
	"ghgo/internal/vocab"
)

type createOrganizationRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type organizationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func newOrganizationResponse(organization *domain.Organization) organizationResponse {
	if organization == nil {
		return organizationResponse{}
	}
	return organizationResponse{
		ID:        organization.ID,
		Name:      organization.Name,
		CreatedAt: formatTime(organization.CreatedAt),
		UpdatedAt: formatTime(organization.UpdatedAt),
	}
}

func newOrganizationResponses(organizations []domain.Organization) []organizationResponse {
	responses := make([]organizationResponse, 0, len(organizations))
	for i := range organizations {
		responses = append(responses, newOrganizationResponse(&organizations[i]))
	}
	return responses
}

type createFacilityRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}

type facilityResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	CountryCode    string `json:"country_code"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func newFacilityResponse(facility *domain.Facility) facilityResponse {
	if facility == nil {
		return facilityResponse{}
	}
	return facilityResponse{
		ID:             facility.ID,
		OrganizationID: facility.OrganizationID,
		Name:           facility.Name,
		CountryCode:    facility.CountryCode,
		CreatedAt:      formatTime(facility.CreatedAt),
		UpdatedAt:      formatTime(facility.UpdatedAt),
	}
}

func newFacilityResponses(facilities []domain.Facility) []facilityResponse {
	responses := make([]facilityResponse, 0, len(facilities))
	for i := range facilities {
		responses = append(responses, newFacilityResponse(&facilities[i]))
	}
	return responses
}

type createReportingPeriodRequest struct {
	ID       string                       `json:"id"`
	Year     int                          `json:"year"`
	StartsOn string                       `json:"starts_on"`
	EndsOn   string                       `json:"ends_on"`
	Status   domain.ReportingPeriodStatus `json:"status"`
}

type reportingPeriodResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Year           int    `json:"year"`
	StartsOn       string `json:"starts_on"`
	EndsOn         string `json:"ends_on"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func newReportingPeriodResponse(period *domain.ReportingPeriod) reportingPeriodResponse {
	if period == nil {
		return reportingPeriodResponse{}
	}
	return reportingPeriodResponse{
		ID:             period.ID,
		OrganizationID: period.OrganizationID,
		Year:           period.Year,
		StartsOn:       formatTime(period.StartsOn),
		EndsOn:         formatTime(period.EndsOn),
		Status:         string(period.Status),
		CreatedAt:      formatTime(period.CreatedAt),
		UpdatedAt:      formatTime(period.UpdatedAt),
	}
}

func newReportingPeriodResponses(periods []domain.ReportingPeriod) []reportingPeriodResponse {
	responses := make([]reportingPeriodResponse, 0, len(periods))
	for i := range periods {
		responses = append(responses, newReportingPeriodResponse(&periods[i]))
	}
	return responses
}

type upsertReportingPeriodSettingsRequest struct {
	ID             string              `json:"id"`
	OrganizationID string              `json:"organization_id"`
	MobileMethod   domain.MobileMethod `json:"mobile_method"`
}

type reportingPeriodSettingsResponse struct {
	ID                string `json:"id"`
	OrganizationID    string `json:"organization_id"`
	ReportingPeriodID string `json:"reporting_period_id"`
	MobileMethod      string `json:"mobile_method"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func newReportingPeriodSettingsResponse(settings *domain.ReportingPeriodSettings) reportingPeriodSettingsResponse {
	if settings == nil {
		return reportingPeriodSettingsResponse{}
	}
	return reportingPeriodSettingsResponse{
		ID:                settings.ID,
		OrganizationID:    settings.OrganizationID,
		ReportingPeriodID: settings.ReportingPeriodID,
		MobileMethod:      string(settings.MobileMethod),
		CreatedAt:         formatTime(settings.CreatedAt),
		UpdatedAt:         formatTime(settings.UpdatedAt),
	}
}

type upsertElectricitySettingsRequest struct {
	ID                    string            `json:"id"`
	OrganizationID        string            `json:"organization_id"`
	HasGuaranteesOfOrigin bool              `json:"has_guarantees_of_origin"`
	GOCoverage            domain.GOCoverage `json:"go_coverage"`
	GOReference           string            `json:"go_reference"`
	GOMarket              string            `json:"go_market"`
	GOCancelledAt         *string           `json:"go_cancelled_at"`
	EvidenceFileID        string            `json:"evidence_file_id"`
}

type electricitySettingsResponse struct {
	ID                    string  `json:"id"`
	OrganizationID        string  `json:"organization_id"`
	ReportingPeriodID     string  `json:"reporting_period_id"`
	FacilityID            string  `json:"facility_id"`
	HasGuaranteesOfOrigin bool    `json:"has_guarantees_of_origin"`
	GOCoverage            string  `json:"go_coverage"`
	GOReference           string  `json:"go_reference"`
	GOMarket              string  `json:"go_market"`
	GOCancelledAt         *string `json:"go_cancelled_at"`
	EvidenceFileID        *string `json:"evidence_file_id"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

func newElectricitySettingsResponse(settings *domain.ElectricitySettings) electricitySettingsResponse {
	if settings == nil {
		return electricitySettingsResponse{}
	}
	return electricitySettingsResponse{
		ID:                    settings.ID,
		OrganizationID:        settings.OrganizationID,
		ReportingPeriodID:     settings.ReportingPeriodID,
		FacilityID:            settings.FacilityID,
		HasGuaranteesOfOrigin: settings.HasGuaranteesOfOrigin,
		GOCoverage:            string(settings.GOCoverage),
		GOReference:           settings.GOReference,
		GOMarket:              settings.GOMarket,
		GOCancelledAt:         formatTimePtr(settings.GOCancelledAt),
		EvidenceFileID:        formatIDPtr(settings.EvidenceFileID),
		CreatedAt:             formatTime(settings.CreatedAt),
		UpdatedAt:             formatTime(settings.UpdatedAt),
	}
}

type parseInputRequest struct {
	InputKind vocab.InputKind `json:"input_kind"`
	RawText   string          `json:"raw_text"`
}

type commitInputRequest struct {
	Context commitContextPayload `json:"context"`
	RawText string               `json:"raw_text"`
}

type commitContextPayload struct {
	OrganizationID        string              `json:"organization_id"`
	ReportingPeriodID     string              `json:"reporting_period_id"`
	FacilityID            *string             `json:"facility_id"`
	InputKind             vocab.InputKind     `json:"input_kind"`
	MobileMethod          domain.MobileMethod `json:"mobile_method"`
	HasGuaranteesOfOrigin bool                `json:"has_guarantees_of_origin"`
	GOCoverage            domain.GOCoverage   `json:"go_coverage"`
}

func (p commitContextPayload) toInput() (input.CommitContext, error) {
	return input.CommitContext{
		OrganizationID:        p.OrganizationID,
		ReportingPeriodID:     p.ReportingPeriodID,
		FacilityID:            p.FacilityID,
		InputKind:             p.InputKind,
		MobileMethod:          p.MobileMethod,
		HasGuaranteesOfOrigin: p.HasGuaranteesOfOrigin,
		GOCoverage:            p.GOCoverage,
	}, nil
}

type parseResultPayload struct {
	InputKind string             `json:"input_kind"`
	RawText   string             `json:"raw_text"`
	Rows      []parsedRowPayload `json:"rows"`
	RowsTotal int                `json:"rows_total"`
	RowsValid int                `json:"rows_valid"`
	RowsError int                `json:"rows_error"`
}

func newParseResultPayload(result input.ParseResult) parseResultPayload {
	rows := make([]parsedRowPayload, 0, len(result.Rows))
	for _, row := range result.Rows {
		rows = append(rows, newParsedRowPayload(row))
	}
	return parseResultPayload{
		InputKind: result.InputKind,
		RawText:   result.RawText,
		Rows:      rows,
		RowsTotal: result.RowsTotal,
		RowsValid: result.RowsValid,
		RowsError: result.RowsError,
	}
}

func (p parseResultPayload) toInput() input.ParseResult {
	rows := make([]input.ParsedRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.toInput())
	}
	return input.ParseResult{
		InputKind: p.InputKind,
		RawText:   p.RawText,
		Rows:      rows,
		RowsTotal: p.RowsTotal,
		RowsValid: p.RowsValid,
		RowsError: p.RowsError,
	}
}

type parsedRowPayload struct {
	RowNumber  int                 `json:"row_number"`
	RawFields  []string            `json:"raw_fields"`
	Normalized map[string]string   `json:"normalized"`
	Errors     []parseIssuePayload `json:"errors"`
	Warnings   []parseIssuePayload `json:"warnings"`
}

func newParsedRowPayload(row input.ParsedRow) parsedRowPayload {
	errors := make([]parseIssuePayload, 0, len(row.Errors))
	for _, issue := range row.Errors {
		errors = append(errors, newParseIssuePayload(issue))
	}
	warnings := make([]parseIssuePayload, 0, len(row.Warnings))
	for _, issue := range row.Warnings {
		warnings = append(warnings, newParseIssuePayload(issue))
	}
	normalized := map[string]string{}
	for key, value := range row.Normalized {
		normalized[key] = value
	}
	return parsedRowPayload{
		RowNumber:  row.RowNumber,
		RawFields:  append([]string(nil), row.RawFields...),
		Normalized: normalized,
		Errors:     errors,
		Warnings:   warnings,
	}
}

func (p parsedRowPayload) toInput() input.ParsedRow {
	errors := make([]input.ParseIssue, 0, len(p.Errors))
	for _, issue := range p.Errors {
		errors = append(errors, issue.toInput())
	}
	warnings := make([]input.ParseIssue, 0, len(p.Warnings))
	for _, issue := range p.Warnings {
		warnings = append(warnings, issue.toInput())
	}
	normalized := map[string]string{}
	for key, value := range p.Normalized {
		normalized[key] = value
	}
	return input.ParsedRow{
		RowNumber:  p.RowNumber,
		RawFields:  append([]string(nil), p.RawFields...),
		Normalized: normalized,
		Errors:     errors,
		Warnings:   warnings,
	}
}

type parseIssuePayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newParseIssuePayload(issue input.ParseIssue) parseIssuePayload {
	return parseIssuePayload{Code: issue.Code, Message: issue.Message}
}

func (p parseIssuePayload) toInput() input.ParseIssue {
	return input.ParseIssue{Code: p.Code, Message: p.Message}
}

type commitResultResponse struct {
	PasteBatchID      string   `json:"paste_batch_id"`
	ActivityRecordIDs []string `json:"activity_record_ids"`
	RowsTotal         int      `json:"rows_total"`
	RowsValid         int      `json:"rows_valid"`
	RowsError         int      `json:"rows_error"`
	Committed         bool     `json:"committed"`
}

func newCommitResultResponse(result input.CommitResult) commitResultResponse {
	return commitResultResponse{
		PasteBatchID:      result.PasteBatchID,
		ActivityRecordIDs: append([]string(nil), result.ActivityRecordIDs...),
		RowsTotal:         result.RowsTotal,
		RowsValid:         result.RowsValid,
		RowsError:         result.RowsError,
		Committed:         result.Committed,
	}
}

type runCalculationRequest struct {
	OrganizationID string `json:"organization_id"`
	FactorSetID    string `json:"factor_set_id"`
}

type calculationResultResponse struct {
	CalculationRunID         string   `json:"calculation_run_id"`
	ResultsCreated           int      `json:"results_created"`
	PrimaryTotalKgCO2e       float64  `json:"primary_total_kg_co2e"`
	LocationBasedTotalKgCO2e *float64 `json:"location_based_total_kg_co2e"`
}

func newCalculationResultResponse(result calc.RunResult) calculationResultResponse {
	return calculationResultResponse{
		CalculationRunID:         result.CalculationRunID,
		ResultsCreated:           result.ResultsCreated,
		PrimaryTotalKgCO2e:       result.PrimaryTotalKgCO2e,
		LocationBasedTotalKgCO2e: result.LocationBasedTotalKgCO2e,
	}
}

type calculationRunResponse struct {
	ID                   string  `json:"id"`
	OrganizationID       string  `json:"organization_id"`
	ReportingPeriodID    string  `json:"reporting_period_id"`
	FactorSetID          string  `json:"factor_set_id"`
	Status               string  `json:"status"`
	StartedAt            string  `json:"started_at"`
	CompletedAt          *string `json:"completed_at"`
	SettingsSnapshotJSON string  `json:"settings_snapshot_json"`
}

func newCalculationRunResponse(run *domain.CalculationRun) calculationRunResponse {
	if run == nil {
		return calculationRunResponse{}
	}
	return calculationRunResponse{
		ID:                   run.ID,
		OrganizationID:       run.OrganizationID,
		ReportingPeriodID:    run.ReportingPeriodID,
		FactorSetID:          run.FactorSetID,
		Status:               string(run.Status),
		StartedAt:            formatTime(run.StartedAt),
		CompletedAt:          formatTimePtr(run.CompletedAt),
		SettingsSnapshotJSON: run.SettingsSnapshotJSON,
	}
}

func newCalculationRunResponses(runs []domain.CalculationRun) []calculationRunResponse {
	responses := make([]calculationRunResponse, 0, len(runs))
	for i := range runs {
		responses = append(responses, newCalculationRunResponse(&runs[i]))
	}
	return responses
}

type factorSetResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Source       string `json:"source"`
	Year         int    `json:"year"`
	Version      string `json:"version"`
	ImportedAt   string `json:"imported_at"`
	MetadataJSON string `json:"metadata_json"`
}

func newFactorSetResponse(factorSet *domain.FactorSet) factorSetResponse {
	if factorSet == nil {
		return factorSetResponse{}
	}
	return factorSetResponse{
		ID:           factorSet.ID,
		Name:         factorSet.Name,
		Source:       factorSet.Source,
		Year:         factorSet.Year,
		Version:      factorSet.Version,
		ImportedAt:   formatTime(factorSet.ImportedAt),
		MetadataJSON: factorSet.MetadataJSON,
	}
}

func newFactorSetResponses(factorSets []domain.FactorSet) []factorSetResponse {
	responses := make([]factorSetResponse, 0, len(factorSets))
	for i := range factorSets {
		responses = append(responses, newFactorSetResponse(&factorSets[i]))
	}
	return responses
}

type reportTablesResponse struct {
	CalculationRunID   string                         `json:"calculation_run_id"`
	OrganizationID     string                         `json:"organization_id"`
	ReportingPeriodID  string                         `json:"reporting_period_id"`
	FactorSetID        string                         `json:"factor_set_id"`
	ExecutiveSummary   executiveSummaryResponse       `json:"executive_summary"`
	MonthlyEmissions   report.MonthlyEmissionsTable   `json:"monthly_emissions"`
	ElectricityDetail  report.ElectricityDetailTable  `json:"electricity_detail"`
	NaturalGasDetail   report.NaturalGasDetailTable   `json:"natural_gas_detail"`
	MobileDetail       report.MobileDetailTable       `json:"mobile_detail"`
	RefrigerantsDetail report.RefrigerantsDetailTable `json:"refrigerants_detail"`
	ScopeSummary       report.ScopeSummaryTable       `json:"scope_summary"`
	VectorSummary      report.VectorSummaryTable      `json:"vector_summary"`
	Methodology        report.MethodologyTable        `json:"methodology"`
	ValidationNotes    report.ValidationNotesTable    `json:"validation_notes"`
}

type executiveSummaryResponse struct {
	ReportingPeriodID          string   `json:"reporting_period_id"`
	FactorSetID                string   `json:"factor_set_id"`
	PrimaryTotalKgCO2e         float64  `json:"primary_total_kg_co2e"`
	PrimaryTotalTCO2e          float64  `json:"primary_total_t_co2e"`
	LocationBasedTotalKgCO2e   *float64 `json:"location_based_total_kg_co2e"`
	LocationBasedTotalTCO2e    *float64 `json:"location_based_total_t_co2e"`
	ElectricityPrimaryMethod   string   `json:"electricity_primary_method"`
	HasLocationBasedComparison bool     `json:"has_location_based_comparison"`
	Scope1PrimaryKgCO2e        float64  `json:"scope1_primary_kg_co2e"`
	Scope2PrimaryKgCO2e        float64  `json:"scope2_primary_kg_co2e"`
	Scope2LocationBasedKgCO2e  *float64 `json:"scope2_location_based_kg_co2e"`
}

func newReportTablesResponse(tables *report.ReportTables) reportTablesResponse {
	if tables == nil {
		return reportTablesResponse{}
	}
	return reportTablesResponse{
		CalculationRunID:   tables.CalculationRunID,
		OrganizationID:     tables.OrganizationID,
		ReportingPeriodID:  tables.ReportingPeriodID,
		FactorSetID:        tables.FactorSetID,
		ExecutiveSummary:   newExecutiveSummaryResponse(tables.ExecutiveSummary),
		MonthlyEmissions:   tables.MonthlyEmissions,
		ElectricityDetail:  tables.ElectricityDetail,
		NaturalGasDetail:   tables.NaturalGasDetail,
		MobileDetail:       tables.MobileDetail,
		RefrigerantsDetail: tables.RefrigerantsDetail,
		ScopeSummary:       tables.ScopeSummary,
		VectorSummary:      tables.VectorSummary,
		Methodology:        tables.Methodology,
		ValidationNotes:    tables.ValidationNotes,
	}
}

func newExecutiveSummaryResponse(summary report.ExecutiveSummaryTable) executiveSummaryResponse {
	return executiveSummaryResponse{
		ReportingPeriodID:          summary.ReportingPeriodID,
		FactorSetID:                summary.FactorSetID,
		PrimaryTotalKgCO2e:         summary.PrimaryTotalKgCO2e,
		PrimaryTotalTCO2e:          summary.PrimaryTotalTCO2e,
		LocationBasedTotalKgCO2e:   summary.LocationBasedTotalKgCO2e,
		LocationBasedTotalTCO2e:    summary.LocationBasedTotalTCO2e,
		ElectricityPrimaryMethod:   summary.ElectricityPrimaryMethod,
		HasLocationBasedComparison: summary.HasLocationBasedComparison,
		Scope1PrimaryKgCO2e:        summary.Scope1PrimaryKgCO2e,
		Scope2PrimaryKgCO2e:        summary.Scope2PrimaryKgCO2e,
		Scope2LocationBasedKgCO2e:  summary.Scope2LocationBasedKgCO2e,
	}
}

func parseOptionalTime(value string, field string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("%s must be an RFC3339 timestamp or YYYY-MM-DD date", field)
}

func parseOptionalTimePtr(value *string, field string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := parseOptionalTime(*value, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatTime(*value)
	return &formatted
}

func formatIDPtr(value *domain.ID) *string {
	if value == nil {
		return nil
	}
	formatted := string(*value)
	return &formatted
}
