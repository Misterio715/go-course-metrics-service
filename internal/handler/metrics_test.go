package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
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

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"temperature","type":"gauge","value":12.5}`))
	req.Header.Set("Content-Type", "application/json")
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
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", rec.Header().Get("Content-Type"))
	}

	var updated models.Metrics
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("expected JSON update response, got error: %v", err)
	}
	if updated.ID != "temperature" || updated.MType != models.Gauge || updated.Value == nil || *updated.Value != 12.5 {
		t.Fatalf("expected updated gauge metric, got %+v", updated)
	}

	valueRequest := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(`{"id":"temperature","type":"gauge"}`))
	valueRequest.Header.Set("Content-Type", "application/json")
	valueRecorder := httptest.NewRecorder()
	router.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", valueRecorder.Header().Get("Content-Type"))
	}

	var stored models.Metrics
	if err := json.NewDecoder(valueRecorder.Body).Decode(&stored); err != nil {
		t.Fatalf("expected JSON value response, got error: %v", err)
	}
	if stored.ID != "temperature" || stored.MType != models.Gauge || stored.Value == nil || *stored.Value != 12.5 {
		t.Fatalf("expected stored gauge metric, got %+v", stored)
	}

	legacyUpdateRequest := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/13.5", nil)
	legacyUpdateRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyUpdateRecorder, legacyUpdateRequest)

	if legacyUpdateRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy update status %d, got %d", http.StatusOK, legacyUpdateRecorder.Code)
	}
	if service.gaugeValue != 13.5 {
		t.Fatalf("expected legacy gauge value %v, got %v", 13.5, service.gaugeValue)
	}

	legacyValueRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	legacyValueRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyValueRecorder, legacyValueRequest)

	if legacyValueRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy value status %d, got %d", http.StatusOK, legacyValueRecorder.Code)
	}
	if legacyValueRecorder.Body.String() != "13.5" {
		t.Fatalf("expected legacy gauge value %q, got %q", "13.5", legacyValueRecorder.Body.String())
	}
}

func TestMetricsHandlerStoresCounterFromRoute(t *testing.T) {
	service := &metricsServiceStub{
		gauges:   map[string]float64{"temperature": 12.5},
		counters: map[string]int64{"requests": 2},
	}
	router := gin.New()
	NewMetricsHandler(service).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"requests","type":"counter","delta":3}`))
	req.Header.Set("Content-Type", "application/json")
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
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", rec.Header().Get("Content-Type"))
	}

	var updated models.Metrics
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("expected JSON update response, got error: %v", err)
	}
	if updated.ID != "requests" || updated.MType != models.Counter || updated.Delta == nil || *updated.Delta != 5 {
		t.Fatalf("expected accumulated counter metric, got %+v", updated)
	}

	valueRequest := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(`{"id":"requests","type":"counter"}`))
	valueRequest.Header.Set("Content-Type", "application/json")
	valueRecorder := httptest.NewRecorder()
	router.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", valueRecorder.Header().Get("Content-Type"))
	}

	var stored models.Metrics
	if err := json.NewDecoder(valueRecorder.Body).Decode(&stored); err != nil {
		t.Fatalf("expected JSON value response, got error: %v", err)
	}
	if stored.ID != "requests" || stored.MType != models.Counter || stored.Delta == nil || *stored.Delta != 5 {
		t.Fatalf("expected stored counter metric, got %+v", stored)
	}

	legacyUpdateRequest := httptest.NewRequest(http.MethodPost, "/update/counter/requests/2", nil)
	legacyUpdateRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyUpdateRecorder, legacyUpdateRequest)

	if legacyUpdateRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy update status %d, got %d", http.StatusOK, legacyUpdateRecorder.Code)
	}
	if service.counterValue != 2 {
		t.Fatalf("expected legacy counter delta %d, got %d", 2, service.counterValue)
	}

	legacyValueRequest := httptest.NewRequest(http.MethodGet, "/value/counter/requests", nil)
	legacyValueRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyValueRecorder, legacyValueRequest)

	if legacyValueRecorder.Code != http.StatusOK {
		t.Fatalf("expected legacy value status %d, got %d", http.StatusOK, legacyValueRecorder.Code)
	}
	if legacyValueRecorder.Body.String() != "7" {
		t.Fatalf("expected legacy counter value %q, got %q", "7", legacyValueRecorder.Body.String())
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

func TestMetricsHandlerRejectsInvalidJSONMetrics(t *testing.T) {
	service := &metricsServiceStub{}
	router := gin.New()
	NewMetricsHandler(service).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"temperature","type":"gauge"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", rec.Header().Get("Content-Type"))
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(`{"id":"unknown","type":"gauge"}`))
	unknownRequest.Header.Set("Content-Type", "application/json")
	unknownRecorder := httptest.NewRecorder()
	router.ServeHTTP(unknownRecorder, unknownRequest)

	if unknownRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, unknownRecorder.Code)
	}
	if unknownRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", unknownRecorder.Header().Get("Content-Type"))
	}

	unknownTypeRequest := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(`{"id":"metric","type":"unknown"}`))
	unknownTypeRequest.Header.Set("Content-Type", "application/json")
	unknownTypeRecorder := httptest.NewRecorder()
	router.ServeHTTP(unknownTypeRecorder, unknownTypeRequest)

	if unknownTypeRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, unknownTypeRecorder.Code)
	}
	if unknownTypeRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", unknownTypeRecorder.Header().Get("Content-Type"))
	}

	legacyInvalidRequest := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/not-a-number", nil)
	legacyInvalidRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyInvalidRecorder, legacyInvalidRequest)

	if legacyInvalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected legacy invalid update status %d, got %d", http.StatusBadRequest, legacyInvalidRecorder.Code)
	}

	legacyUnknownRequest := httptest.NewRequest(http.MethodGet, "/value/gauge/unknown", nil)
	legacyUnknownRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyUnknownRecorder, legacyUnknownRequest)

	if legacyUnknownRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected legacy unknown value status %d, got %d", http.StatusNotFound, legacyUnknownRecorder.Code)
	}

	legacyUnknownTypeRequest := httptest.NewRequest(http.MethodGet, "/value/unknown/metric", nil)
	legacyUnknownTypeRecorder := httptest.NewRecorder()
	router.ServeHTTP(legacyUnknownTypeRecorder, legacyUnknownTypeRequest)

	if legacyUnknownTypeRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected legacy unknown type status %d, got %d", http.StatusNotFound, legacyUnknownTypeRecorder.Code)
	}
}
