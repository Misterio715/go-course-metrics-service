package repository

import "testing"

func TestMemStorageStoresGauge(t *testing.T) {
	storage := NewMemStorage()

	storage.AddGauge("temperature", 12.5)
	storage.AddGauge("temperature", 13.5)

	if storage.Gauges["temperature"] != 13.5 {
		t.Fatalf("expected gauge to be overwritten with %v, got %v", 13.5, storage.Gauges["temperature"])
	}
}

func TestMemStorageAccumulatesCounter(t *testing.T) {
	storage := NewMemStorage()

	storage.AddCounter("requests", 2)
	storage.AddCounter("requests", 3)

	if storage.Counters["requests"] != 5 {
		t.Fatalf("expected counter to accumulate to %d, got %d", 5, storage.Counters["requests"])
	}
}
