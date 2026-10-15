package service

import "testing"

type metricsRepositoryStub struct {
	gaugeName    string
	gaugeValue   float64
	counterName  string
	counterValue int64
}

func (r *metricsRepositoryStub) AddGauge(name string, value float64) {
	r.gaugeName = name
	r.gaugeValue = value
}

func (r *metricsRepositoryStub) AddCounter(name string, value int64) {
	r.counterName = name
	r.counterValue = value
}

func TestMetricsServiceAddsGaugeToRepository(t *testing.T) {
	repository := &metricsRepositoryStub{}
	service := NewMetricsService(repository)

	service.AddGauge("temperature", 12.5)

	if repository.gaugeName != "temperature" {
		t.Fatalf("expected gauge name %q, got %q", "temperature", repository.gaugeName)
	}
	if repository.gaugeValue != 12.5 {
		t.Fatalf("expected gauge value %v, got %v", 12.5, repository.gaugeValue)
	}
}

func TestMetricsServiceAddsCounterToRepository(t *testing.T) {
	repository := &metricsRepositoryStub{}
	service := NewMetricsService(repository)

	service.AddCounter("requests", 3)

	if repository.counterName != "requests" {
		t.Fatalf("expected counter name %q, got %q", "requests", repository.counterName)
	}
	if repository.counterValue != 3 {
		t.Fatalf("expected counter value %d, got %d", 3, repository.counterValue)
	}
}
