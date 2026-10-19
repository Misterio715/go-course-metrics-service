package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
	"resty.dev/v3"
)

func TestSendMetricsPostsKnownMetricsAsJSON(t *testing.T) {
	var gotPaths []string
	var gotMethods []string
	var gotContentTypes []string
	var gotMetrics []models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		gotMethods = append(gotMethods, r.Method)
		gotContentTypes = append(gotContentTypes, r.Header.Get("Content-Type"))

		var metric models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			t.Errorf("expected JSON metric body, got error: %v", err)
		}
		gotMetrics = append(gotMetrics, metric)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseURL := server.URL + "/update"

	counter := 7
	metrics := map[string]any{
		"gauge/Alloc":       uint64(42),
		"gauge/RandomValue": 0.25,
		"counter/PollCount": counter,
	}

	sendMetrics(resty.NewWithClient(server.Client()), baseURL, metrics)

	wantMetrics := map[string]struct {
		metricType string
		value      float64
	}{
		"Alloc":       {metricType: models.Gauge, value: 42},
		"RandomValue": {metricType: models.Gauge, value: 0.25},
		"PollCount":   {metricType: models.Counter, value: 7},
	}

	if len(gotMetrics) != len(wantMetrics) {
		t.Fatalf("expected %d requests, got %d", len(wantMetrics), len(gotMetrics))
	}

	for i, path := range gotPaths {
		if path != "/update" {
			t.Fatalf("expected request path %q, got %q", "/update", path)
		}

		if gotMethods[i] != http.MethodPost {
			t.Fatalf("expected request %q to use POST, got %s", path, gotMethods[i])
		}
		if gotContentTypes[i] != "application/json" {
			t.Fatalf("expected request %q content type application/json, got %q", path, gotContentTypes[i])
		}
	}

	for _, metric := range gotMetrics {
		want, ok := wantMetrics[metric.ID]
		if !ok {
			t.Fatalf("unexpected metric ID %q", metric.ID)
		}
		if metric.MType != want.metricType {
			t.Fatalf("expected metric %q type %q, got %q", metric.ID, want.metricType, metric.MType)
		}

		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil || *metric.Value != want.value {
				t.Fatalf("expected gauge %q value %v, got %v", metric.ID, want.value, metric.Value)
			}
			if metric.Delta != nil {
				t.Fatalf("expected gauge %q delta to be omitted, got %d", metric.ID, *metric.Delta)
			}
		case models.Counter:
			if metric.Delta == nil || *metric.Delta != int64(want.value) {
				t.Fatalf("expected counter %q delta %v, got %v", metric.ID, want.value, metric.Delta)
			}
			if metric.Value != nil {
				t.Fatalf("expected counter %q value to be omitted, got %v", metric.ID, *metric.Value)
			}
		}

		delete(wantMetrics, metric.ID)
	}

	if len(wantMetrics) != 0 {
		t.Fatalf("expected all metrics to be sent, missing %v", wantMetrics)
	}
}
