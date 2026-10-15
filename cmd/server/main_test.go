package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerHandlerAcceptsMetricUpdates(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/12.5", nil)
	rec := httptest.NewRecorder()

	newServerHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
