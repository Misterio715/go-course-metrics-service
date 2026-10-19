package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
)

func TestNewServerHandlerAcceptsMetricUpdates(t *testing.T) {
	server := newServerHandler()
	updateRequest := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"temperature","type":"gauge","value":12.5}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRecorder := httptest.NewRecorder()

	server.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRecorder.Code)
	}
	if updateRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", updateRecorder.Header().Get("Content-Type"))
	}

	valueRequest := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(`{"id":"temperature","type":"gauge"}`))
	valueRequest.Header.Set("Content-Type", "application/json")
	valueRecorder := httptest.NewRecorder()
	server.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", valueRecorder.Header().Get("Content-Type"))
	}

	var metric models.Metrics
	if err := json.NewDecoder(valueRecorder.Body).Decode(&metric); err != nil {
		t.Fatalf("expected JSON value response, got error: %v", err)
	}
	if metric.ID != "temperature" || metric.MType != models.Gauge || metric.Value == nil || *metric.Value != 12.5 {
		t.Fatalf("expected stored gauge metric, got %+v", metric)
	}

	legacyUpdateRequest := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/13.5", nil)
	legacyUpdateRecorder := httptest.NewRecorder()
	server.ServeHTTP(legacyUpdateRecorder, legacyUpdateRequest)

	if legacyUpdateRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy update status %d, got %d", http.StatusOK, legacyUpdateRecorder.Code)
	}

	legacyValueRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	legacyValueRecorder := httptest.NewRecorder()
	server.ServeHTTP(legacyValueRecorder, legacyValueRequest)

	if legacyValueRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy value status %d, got %d", http.StatusOK, legacyValueRecorder.Code)
	}
	if legacyValueRecorder.Body.String() != "13.5" {
		t.Fatalf("expected legacy metric value %q, got %q", "13.5", legacyValueRecorder.Body.String())
	}

	pageRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	pageRecorder := httptest.NewRecorder()
	server.ServeHTTP(pageRecorder, pageRequest)

	if pageRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageRecorder.Code)
	}
	if !strings.Contains(pageRecorder.Body.String(), "temperature") {
		t.Fatalf("expected page to contain metric name, got %q", pageRecorder.Body.String())
	}

	unknownTypeRequest := httptest.NewRequest(http.MethodPost, "/update/unknown/testCounter/100", nil)
	unknownTypeRecorder := httptest.NewRecorder()
	server.ServeHTTP(unknownTypeRecorder, unknownTypeRequest)

	if unknownTypeRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, unknownTypeRecorder.Code)
	}
}
