package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ghgo/internal/app"
	"ghgo/internal/vocab"
)

func TestOrganizationsCreateListGet(t *testing.T) {
	handler := newTestHandler(t)

	status, envelope := requestJSON(t, handler, http.MethodPost, "/organizations", createOrganizationRequest{Name: "Acme Ltd"})
	if status != http.StatusCreated {
		t.Fatalf("create organization status = %d, want %d: %s", status, http.StatusCreated, envelopeText(envelope))
	}
	var created organizationResponse
	decodeData(t, envelope, &created)
	if created.ID == "" || created.Name != "Acme Ltd" {
		t.Fatalf("created organization = %#v, want id and name", created)
	}

	status, envelope = requestJSON(t, handler, http.MethodGet, "/organizations", nil)
	if status != http.StatusOK {
		t.Fatalf("list organizations status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var organizations []organizationResponse
	decodeData(t, envelope, &organizations)
	if len(organizations) != 1 || organizations[0].ID != created.ID {
		t.Fatalf("organizations = %#v, want created organization", organizations)
	}

	status, envelope = requestJSON(t, handler, http.MethodGet, "/organizations/"+created.ID, nil)
	if status != http.StatusOK {
		t.Fatalf("get organization status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var got organizationResponse
	decodeData(t, envelope, &got)
	if got != created {
		t.Fatalf("organization = %#v, want %#v", got, created)
	}
}

func TestAPIErrorMapping(t *testing.T) {
	handler := newTestHandler(t)

	status, envelope := requestJSON(t, handler, http.MethodGet, "/organizations/missing", nil)
	if status != http.StatusNotFound {
		t.Fatalf("missing organization status = %d, want %d: %s", status, http.StatusNotFound, envelopeText(envelope))
	}
	if envelope.Error == nil || envelope.Error.Code != "not_found" {
		t.Fatalf("missing organization error = %#v, want not_found", envelope.Error)
	}

	status, envelope = requestRaw(t, handler, http.MethodPost, "/organizations", "{")
	if status != http.StatusBadRequest {
		t.Fatalf("invalid JSON status = %d, want %d: %s", status, http.StatusBadRequest, envelopeText(envelope))
	}
	if envelope.Error == nil || envelope.Error.Code != "malformed_json" {
		t.Fatalf("invalid JSON error = %#v, want malformed_json", envelope.Error)
	}

	status, envelope = requestJSON(t, handler, http.MethodPut, "/organizations", createOrganizationRequest{Name: "Acme Ltd"})
	if status != http.StatusMethodNotAllowed {
		t.Fatalf("unsupported method status = %d, want %d: %s", status, http.StatusMethodNotAllowed, envelopeText(envelope))
	}
	if envelope.Error == nil || envelope.Error.Code != "method_not_allowed" {
		t.Fatalf("unsupported method error = %#v, want method_not_allowed", envelope.Error)
	}
}

func TestNaturalGasWorkflow(t *testing.T) {
	handler := newTestHandler(t)
	organization := createTestOrganization(t, handler)
	facility := createTestFacility(t, handler, organization.ID)
	period := createTestReportingPeriod(t, handler, organization.ID)

	status, envelope := requestJSON(t, handler, http.MethodPost, "/inputs/parse", parseInputRequest{
		InputKind: vocab.InputNaturalGasMonthlySmc,
		RawText:   "Month\tConsumption\nJanuary\t100",
	})
	if status != http.StatusOK {
		t.Fatalf("parse status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var parsed parseResultPayload
	decodeData(t, envelope, &parsed)
	if parsed.RowsValid != 1 || parsed.RowsError != 0 {
		t.Fatalf("parsed result = %#v, want one valid row", parsed)
	}

	facilityID := facility.ID
	status, envelope = requestJSON(t, handler, http.MethodPost, "/inputs/commit", commitInputRequest{
		Context: commitContextPayload{
			OrganizationID:    organization.ID,
			ReportingPeriodID: period.ID,
			FacilityID:        &facilityID,
			ReportingYear:     period.Year,
			PeriodStart:       period.StartsOn,
			PeriodEnd:         period.EndsOn,
			InputKind:         vocab.InputNaturalGasMonthlySmc,
		},
		Parsed: parsed,
	})
	if status != http.StatusOK {
		t.Fatalf("commit status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var commit commitResultResponse
	decodeData(t, envelope, &commit)
	if !commit.Committed || len(commit.ActivityRecordIDs) != 1 {
		t.Fatalf("commit = %#v, want one committed activity record", commit)
	}

	status, envelope = requestJSON(t, handler, http.MethodPost, "/reporting-periods/"+period.ID+"/calculations", runCalculationRequest{OrganizationID: organization.ID})
	if status != http.StatusCreated {
		t.Fatalf("calculation status = %d, want %d: %s", status, http.StatusCreated, envelopeText(envelope))
	}
	var calculation calculationResultResponse
	decodeData(t, envelope, &calculation)
	if calculation.CalculationRunID == "" || calculation.PrimaryTotalKgCO2e <= 0 {
		t.Fatalf("calculation = %#v, want run id and positive emissions", calculation)
	}

	status, envelope = requestJSON(t, handler, http.MethodGet, "/reporting-periods/"+period.ID+"/calculation-runs", nil)
	if status != http.StatusOK {
		t.Fatalf("list calculation runs status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var runs []calculationRunResponse
	decodeData(t, envelope, &runs)
	if len(runs) != 1 || runs[0].ID != calculation.CalculationRunID {
		t.Fatalf("calculation runs = %#v, want created run", runs)
	}

	status, envelope = requestJSON(t, handler, http.MethodGet, "/calculation-runs/"+calculation.CalculationRunID, nil)
	if status != http.StatusOK {
		t.Fatalf("get calculation run status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var run calculationRunResponse
	decodeData(t, envelope, &run)
	if run.ID != calculation.CalculationRunID || run.Status != "completed" {
		t.Fatalf("calculation run = %#v, want completed run", run)
	}

	status, envelope = requestJSON(t, handler, http.MethodGet, "/calculation-runs/"+calculation.CalculationRunID+"/report-tables", nil)
	if status != http.StatusOK {
		t.Fatalf("report tables status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var tables reportTablesResponse
	decodeData(t, envelope, &tables)
	if tables.CalculationRunID != calculation.CalculationRunID || tables.ExecutiveSummary.PrimaryTotalKgCO2e != calculation.PrimaryTotalKgCO2e {
		t.Fatalf("report tables = %#v, want matching calculation total", tables.ExecutiveSummary)
	}
}

func TestParseRowErrorsAreOKAndCommitValidationIsBadRequest(t *testing.T) {
	handler := newTestHandler(t)
	organization := createTestOrganization(t, handler)
	facility := createTestFacility(t, handler, organization.ID)
	period := createTestReportingPeriod(t, handler, organization.ID)

	status, envelope := requestJSON(t, handler, http.MethodPost, "/inputs/parse", parseInputRequest{
		InputKind: vocab.InputNaturalGasMonthlySmc,
		RawText:   "NotAMonth\t100",
	})
	if status != http.StatusOK {
		t.Fatalf("parse invalid rows status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var parsedWithErrors parseResultPayload
	decodeData(t, envelope, &parsedWithErrors)
	if parsedWithErrors.RowsError == 0 {
		t.Fatalf("parsed result = %#v, want row errors", parsedWithErrors)
	}

	facilityID := facility.ID
	status, envelope = requestJSON(t, handler, http.MethodPost, "/inputs/commit", commitInputRequest{
		Context: commitContextPayload{
			OrganizationID:    organization.ID,
			ReportingPeriodID: period.ID,
			FacilityID:        &facilityID,
			ReportingYear:     period.Year,
			PeriodStart:       period.StartsOn,
			PeriodEnd:         period.EndsOn,
			InputKind:         vocab.InputNaturalGasMonthlySmc,
		},
		Parsed: parsedWithErrors,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("commit invalid parsed status = %d, want %d: %s", status, http.StatusBadRequest, envelopeText(envelope))
	}
	if envelope.Error == nil || envelope.Error.Code != "invalid_input_commit" {
		t.Fatalf("commit invalid parsed error = %#v, want invalid_input_commit", envelope.Error)
	}

	status, envelope = requestJSON(t, handler, http.MethodPost, "/inputs/parse", parseInputRequest{
		InputKind: vocab.InputNaturalGasMonthlySmc,
		RawText:   "January\t100",
	})
	if status != http.StatusOK {
		t.Fatalf("parse valid row status = %d, want %d: %s", status, http.StatusOK, envelopeText(envelope))
	}
	var validParsed parseResultPayload
	decodeData(t, envelope, &validParsed)

	status, envelope = requestJSON(t, handler, http.MethodPost, "/inputs/commit", commitInputRequest{Parsed: validParsed})
	if status != http.StatusBadRequest {
		t.Fatalf("commit invalid context status = %d, want %d: %s", status, http.StatusBadRequest, envelopeText(envelope))
	}
	if envelope.Error == nil || envelope.Error.Code != "invalid_input_commit" {
		t.Fatalf("commit invalid context error = %#v, want invalid_input_commit", envelope.Error)
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	backend, err := app.OpenSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open sqlite backend: %v", err)
	}
	t.Cleanup(func() {
		if err := backend.Close(); err != nil {
			t.Fatalf("close backend: %v", err)
		}
	})
	return New(backend.Services)
}

func createTestOrganization(t *testing.T, handler http.Handler) organizationResponse {
	t.Helper()
	status, envelope := requestJSON(t, handler, http.MethodPost, "/organizations", createOrganizationRequest{Name: "Acme Ltd"})
	if status != http.StatusCreated {
		t.Fatalf("create organization status = %d, want %d: %s", status, http.StatusCreated, envelopeText(envelope))
	}
	var organization organizationResponse
	decodeData(t, envelope, &organization)
	return organization
}

func createTestFacility(t *testing.T, handler http.Handler, organizationID string) facilityResponse {
	t.Helper()
	status, envelope := requestJSON(t, handler, http.MethodPost, "/organizations/"+organizationID+"/facilities", createFacilityRequest{
		Name:        "Milan Office",
		CountryCode: "IT",
	})
	if status != http.StatusCreated {
		t.Fatalf("create facility status = %d, want %d: %s", status, http.StatusCreated, envelopeText(envelope))
	}
	var facility facilityResponse
	decodeData(t, envelope, &facility)
	return facility
}

func createTestReportingPeriod(t *testing.T, handler http.Handler, organizationID string) reportingPeriodResponse {
	t.Helper()
	status, envelope := requestJSON(t, handler, http.MethodPost, "/organizations/"+organizationID+"/reporting-periods", createReportingPeriodRequest{Year: 2026})
	if status != http.StatusCreated {
		t.Fatalf("create reporting period status = %d, want %d: %s", status, http.StatusCreated, envelopeText(envelope))
	}
	var period reportingPeriodResponse
	decodeData(t, envelope, &period)
	return period
}

type testEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *errorObject    `json:"error"`
}

func requestJSON(t *testing.T, handler http.Handler, method string, path string, body any) (int, testEnvelope) {
	t.Helper()
	if body == nil {
		return requestRawReader(t, handler, method, path, nil)
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return requestRawReader(t, handler, method, path, bytes.NewReader(data))
}

func requestRaw(t *testing.T, handler http.Handler, method string, path string, body string) (int, testEnvelope) {
	t.Helper()
	return requestRawReader(t, handler, method, path, bytes.NewBufferString(body))
}

func requestRawReader(t *testing.T, handler http.Handler, method string, path string, body io.Reader) (int, testEnvelope) {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var envelope testEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
	return rec.Code, envelope
}

func decodeData(t *testing.T, envelope testEnvelope, dst any) {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("response error = %#v", envelope.Error)
	}
	if len(envelope.Data) == 0 {
		t.Fatalf("response data is empty")
	}
	if err := json.Unmarshal(envelope.Data, dst); err != nil {
		t.Fatalf("decode response data %s: %v", string(envelope.Data), err)
	}
}

func envelopeText(envelope testEnvelope) string {
	if envelope.Error != nil {
		return fmt.Sprintf("error=%s:%s", envelope.Error.Code, envelope.Error.Message)
	}
	return string(envelope.Data)
}
