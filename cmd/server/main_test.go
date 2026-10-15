package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServerHandlerAcceptsMetricUpdates(t *testing.T) {
	server := newServerHandler()
	updateRequest := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/12.5", nil)
	updateRecorder := httptest.NewRecorder()

	server.ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRecorder.Code)
	}

	valueRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	valueRecorder := httptest.NewRecorder()
	server.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Body.String() != "12.5" {
		t.Fatalf("expected metric value %q, got %q", "12.5", valueRecorder.Body.String())
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
