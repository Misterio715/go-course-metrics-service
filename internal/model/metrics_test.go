package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMetricsJSONOmitsUnsetDeltaAndValue(t *testing.T) {
	data, err := json.Marshal(Metrics{ID: "requests", MType: Counter})
	if err != nil {
		t.Fatalf("expected metrics to marshal, got error: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, `"id":"requests"`) {
		t.Fatalf("expected id to be encoded, got %s", body)
	}
	if strings.Contains(body, "delta") {
		t.Fatalf("expected unset delta to be omitted, got %s", body)
	}
	if strings.Contains(body, "value") {
		t.Fatalf("expected unset value to be omitted, got %s", body)
	}
}

func TestMetricTypeConstants(t *testing.T) {
	if Counter != "counter" {
		t.Fatalf("expected Counter constant to be %q, got %q", "counter", Counter)
	}
	if Gauge != "gauge" {
		t.Fatalf("expected Gauge constant to be %q, got %q", "gauge", Gauge)
	}
}
