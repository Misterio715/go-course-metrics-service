package repository

import "testing"

func TestMemStorageStoresGauge(t *testing.T) {
	storage := NewMemStorage()

	storage.AddGauge("temperature", 12.5)
	storage.AddGauge("temperature", 13.5)

	value, ok := storage.GetGauge("temperature")
	if !ok {
		t.Fatal("expected gauge to exist")
	}
	if value != 13.5 {
		t.Fatalf("expected gauge to be overwritten with %v, got %v", 13.5, value)
	}

	if _, ok := storage.GetGauge("unknown"); ok {
		t.Fatal("expected unknown gauge to be absent")
	}
}

func TestMemStorageAccumulatesCounter(t *testing.T) {
	storage := NewMemStorage()

	storage.AddCounter("requests", 2)
	storage.AddCounter("requests", 3)

	value, ok := storage.GetCounter("requests")
	if !ok {
		t.Fatal("expected counter to exist")
	}
	if value != 5 {
		t.Fatalf("expected counter to accumulate to %d, got %d", 5, value)
	}

	gauges, counters := storage.GetMetrics()
	if len(gauges) != 0 {
		t.Fatalf("expected no gauges, got %v", gauges)
	}
	if counters["requests"] != 5 {
		t.Fatalf("expected metrics list to contain counter value %d, got %d", 5, counters["requests"])
	}
}
