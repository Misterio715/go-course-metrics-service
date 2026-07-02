package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type metricsServiceStub struct {
	gaugeName    string
	gaugeValue   float64
	counterName  string
	counterValue int64
}

func (s *metricsServiceStub) AddGauge(name string, value float64) {
	s.gaugeName = name
	s.gaugeValue = value
}

func (s *metricsServiceStub) AddCounter(name string, value int64) {
	s.counterName = name
	s.counterValue = value
}

func TestMetricsHandlerStoresGaugeFromRoute(t *testing.T) {
	service := &metricsServiceStub{}
	mux := http.NewServeMux()
	NewMetricsHandler(service).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/12.5", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if service.gaugeName != "temperature" {
		t.Fatalf("expected gauge name %q, got %q", "temperature", service.gaugeName)
	}
	if service.gaugeValue != 12.5 {
		t.Fatalf("expected gauge value %v, got %v", 12.5, service.gaugeValue)
	}
}

func TestMetricsHandlerStoresCounterFromRoute(t *testing.T) {
	service := &metricsServiceStub{}
	mux := http.NewServeMux()
	NewMetricsHandler(service).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/requests/3", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if service.counterName != "requests" {
		t.Fatalf("expected counter name %q, got %q", "requests", service.counterName)
	}
	if service.counterValue != 3 {
		t.Fatalf("expected counter value %d, got %d", 3, service.counterValue)
	}
}

func TestMetricsHandlerRejectsInvalidGaugeValue(t *testing.T) {
	service := &metricsServiceStub{}
	mux := http.NewServeMux()
	NewMetricsHandler(service).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/not-a-number", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
