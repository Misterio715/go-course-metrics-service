package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type metricsServiceStub struct {
	gaugeName    string
	gaugeValue   float64
	counterName  string
	counterValue int64
	gauges       map[string]float64
	counters     map[string]int64
}

func (s *metricsServiceStub) AddGauge(name string, value float64) {
	s.gaugeName = name
	s.gaugeValue = value
	if s.gauges == nil {
		s.gauges = make(map[string]float64)
	}
	s.gauges[name] = value
}

func (s *metricsServiceStub) AddCounter(name string, value int64) {
	s.counterName = name
	s.counterValue = value
	if s.counters == nil {
		s.counters = make(map[string]int64)
	}
	s.counters[name] += value
}

func (s *metricsServiceStub) GetGauge(name string) (float64, bool) {
	value, ok := s.gauges[name]
	return value, ok
}

func (s *metricsServiceStub) GetCounter(name string) (int64, bool) {
	value, ok := s.counters[name]
	return value, ok
}

func (s *metricsServiceStub) GetMetrics() (map[string]float64, map[string]int64) {
	return s.gauges, s.counters
}

func TestMetricsHandlerStoresGaugeFromRoute(t *testing.T) {
	service := &metricsServiceStub{}
	router := gin.New()
	NewMetricsHandler(service).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/12.5", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if service.gaugeName != "temperature" {
		t.Fatalf("expected gauge name %q, got %q", "temperature", service.gaugeName)
	}
	if service.gaugeValue != 12.5 {
		t.Fatalf("expected gauge value %v, got %v", 12.5, service.gaugeValue)
	}

	valueRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	valueRecorder := httptest.NewRecorder()
	router.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Body.String() != "12.5" {
		t.Fatalf("expected gauge value %q, got %q", "12.5", valueRecorder.Body.String())
	}
}

func TestMetricsHandlerStoresCounterFromRoute(t *testing.T) {
	service := &metricsServiceStub{gauges: map[string]float64{"temperature": 12.5}}
	router := gin.New()
	NewMetricsHandler(service).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/requests/3", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if service.counterName != "requests" {
		t.Fatalf("expected counter name %q, got %q", "requests", service.counterName)
	}
	if service.counterValue != 3 {
		t.Fatalf("expected counter value %d, got %d", 3, service.counterValue)
	}

	valueRequest := httptest.NewRequest(http.MethodGet, "/value/counter/requests", nil)
	valueRecorder := httptest.NewRecorder()
	router.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Body.String() != "3" {
		t.Fatalf("expected counter value %q, got %q", "3", valueRecorder.Body.String())
	}

	pageRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	pageRecorder := httptest.NewRecorder()
	router.ServeHTTP(pageRecorder, pageRequest)

	if pageRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageRecorder.Code)
	}
	if !strings.HasPrefix(pageRecorder.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected HTML content type, got %q", pageRecorder.Header().Get("Content-Type"))
	}
	if !strings.Contains(pageRecorder.Body.String(), "temperature") || !strings.Contains(pageRecorder.Body.String(), "requests") {
		t.Fatalf("expected page to list metrics, got %q", pageRecorder.Body.String())
	}
}

func TestMetricsHandlerRejectsInvalidGaugeValue(t *testing.T) {
	service := &metricsServiceStub{}
	router := gin.New()
	NewMetricsHandler(service).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/not-a-number", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	unknownRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/unknown", nil)
	unknownRecorder := httptest.NewRecorder()
	router.ServeHTTP(unknownRecorder, unknownRequest)

	if unknownRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, unknownRecorder.Code)
	}

	unknownTypeRequest := httptest.NewRequest(http.MethodGet, "/value/unknown/metric", nil)
	unknownTypeRecorder := httptest.NewRecorder()
	router.ServeHTTP(unknownTypeRecorder, unknownTypeRequest)

	if unknownTypeRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, unknownTypeRecorder.Code)
	}
}
