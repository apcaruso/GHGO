package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"ghgo/internal/app"
	"ghgo/internal/calc"
	"ghgo/internal/input"
	"ghgo/internal/ports"
	"ghgo/internal/report"
)

const maxJSONBodyBytes int64 = 1 << 20

type responseEnvelope struct {
	Data  any          `json:"data,omitempty"`
	Error *errorObject `json:"error,omitempty"`
}

type errorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiHandledMarker interface {
	markAPIHandled()
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed_json", malformedJSONMessage(err))
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(w, http.StatusBadRequest, "malformed_json", "request body must contain a single JSON value")
		return false
	}
	return true
}

func malformedJSONMessage(err error) string {
	if errors.Is(err, io.EOF) {
		return "request body must contain JSON"
	}

	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return fmt.Sprintf("request body exceeds %d bytes", maxBytesError.Limit)
	}

	return "malformed JSON: " + err.Error()
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeEnvelope(w, status, responseEnvelope{Data: data})
}

func writeServiceError(w http.ResponseWriter, err error) {
	status, code, message := serviceErrorResponse(err)
	writeAPIError(w, status, code, message)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	writeEnvelope(w, status, responseEnvelope{Error: &errorObject{Code: code, Message: message}})
}

func writeEnvelope(w http.ResponseWriter, status int, envelope responseEnvelope) {
	if marker, ok := w.(apiHandledMarker); ok {
		marker.markAPIHandled()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

func serviceErrorResponse(err error) (int, string, string) {
	if err == nil {
		return http.StatusInternalServerError, "internal_error", "internal server error"
	}

	switch {
	case errors.Is(err, app.ErrInvalidOptions), errors.Is(err, calc.ErrInvalidOptions), errors.Is(err, report.ErrInvalidOptions):
		return http.StatusBadRequest, "invalid_options", err.Error()
	case errors.Is(err, input.ErrInvalidCommit):
		return http.StatusBadRequest, "invalid_input_commit", err.Error()
	case errors.Is(err, ports.ErrNotFound):
		return http.StatusNotFound, "not_found", err.Error()
	case errors.Is(err, calc.ErrNoActiveRecords):
		return http.StatusConflict, "no_active_records", err.Error()
	case errors.Is(err, calc.ErrMissingFactor):
		return http.StatusConflict, "missing_factor", err.Error()
	case errors.Is(err, calc.ErrInvalidSettings):
		return http.StatusConflict, "invalid_calculation_settings", err.Error()
	case errors.Is(err, report.ErrRunNotCompleted):
		return http.StatusConflict, "report_run_not_completed", err.Error()
	case errors.Is(err, report.ErrMixedMobileMethods):
		return http.StatusConflict, "mixed_mobile_methods", err.Error()
	}

	return http.StatusInternalServerError, "internal_error", "internal server error"
}

func withJSONFallback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture := newCaptureResponse()
		next.ServeHTTP(capture, r)

		if !capture.handled {
			switch capture.statusCode() {
			case http.StatusMethodNotAllowed:
				if allow := capture.Header().Get("Allow"); allow != "" {
					w.Header().Set("Allow", allow)
				}
				writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
				return
			case http.StatusNotFound:
				writeAPIError(w, http.StatusNotFound, "not_found", "not found")
				return
			}
		}

		capture.flushTo(w)
	})
}

type captureResponse struct {
	header      http.Header
	body        bytes.Buffer
	status      int
	wroteHeader bool
	handled     bool
}

func newCaptureResponse() *captureResponse {
	return &captureResponse{header: make(http.Header)}
}

func (r *captureResponse) Header() http.Header {
	return r.header
}

func (r *captureResponse) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
}

func (r *captureResponse) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(p)
}

func (r *captureResponse) markAPIHandled() {
	r.handled = true
}

func (r *captureResponse) statusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *captureResponse) flushTo(w http.ResponseWriter) {
	for name, values := range r.header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(r.statusCode())
	_, _ = r.body.WriteTo(w)
}
