package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
)

func TestNewServerHandlerAcceptsMetricUpdates(t *testing.T) {
	server := newServerHandler()

	var compressedUpdate bytes.Buffer
	compressor := gzip.NewWriter(&compressedUpdate)
	if _, err := compressor.Write([]byte(`{"id":"temperature","type":"gauge","value":12.5}`)); err != nil {
		t.Fatalf("compress update request: %v", err)
	}
	if err := compressor.Close(); err != nil {
		t.Fatalf("close update request compressor: %v", err)
	}

	updateRequest := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(compressedUpdate.Bytes()))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set("Content-Encoding", "gzip")
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
	valueRequest.Header.Set("Accept-Encoding", "gzip")
	valueRecorder := httptest.NewRecorder()
	server.ServeHTTP(valueRecorder, valueRequest)

	if valueRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, valueRecorder.Code)
	}
	if valueRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", valueRecorder.Header().Get("Content-Type"))
	}
	if valueRecorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", valueRecorder.Header().Get("Content-Encoding"))
	}

	valueReader, err := gzip.NewReader(valueRecorder.Body)
	if err != nil {
		t.Fatalf("decompress JSON value response: %v", err)
	}
	defer valueReader.Close()
	var metric models.Metrics
	if err := json.NewDecoder(valueReader).Decode(&metric); err != nil {
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
	pageRequest.Header.Set("Accept-Encoding", "gzip")
	pageRecorder := httptest.NewRecorder()
	server.ServeHTTP(pageRecorder, pageRequest)

	if pageRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageRecorder.Code)
	}
	if pageRecorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", pageRecorder.Header().Get("Content-Encoding"))
	}
	pageReader, err := gzip.NewReader(pageRecorder.Body)
	if err != nil {
		t.Fatalf("decompress metrics page: %v", err)
	}
	defer pageReader.Close()
	pageBody, err := io.ReadAll(pageReader)
	if err != nil {
		t.Fatalf("read metrics page: %v", err)
	}
	if !strings.Contains(string(pageBody), "temperature") {
		t.Fatalf("expected page to contain metric name, got %q", pageBody)
	}

	unknownTypeRequest := httptest.NewRequest(http.MethodPost, "/update/unknown/testCounter/100", nil)
	unknownTypeRecorder := httptest.NewRecorder()
	server.ServeHTTP(unknownTypeRecorder, unknownTypeRequest)

	if unknownTypeRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, unknownTypeRecorder.Code)
	}
}
