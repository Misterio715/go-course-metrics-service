package service

import "testing"

type metricsRepositoryStub struct {
	gaugeName    string
	gaugeValue   float64
	counterName  string
	counterValue int64
	gauges       map[string]float64
	counters     map[string]int64
}

func (r *metricsRepositoryStub) AddGauge(name string, value float64) {
	r.gaugeName = name
	r.gaugeValue = value
	if r.gauges == nil {
		r.gauges = make(map[string]float64)
	}
	r.gauges[name] = value
}

func (r *metricsRepositoryStub) AddCounter(name string, value int64) {
	r.counterName = name
	r.counterValue = value
	if r.counters == nil {
		r.counters = make(map[string]int64)
	}
	r.counters[name] += value
}

func (r *metricsRepositoryStub) GetGauge(name string) (float64, bool) {
	value, ok := r.gauges[name]
	return value, ok
}

func (r *metricsRepositoryStub) GetCounter(name string) (int64, bool) {
	value, ok := r.counters[name]
	return value, ok
}

func (r *metricsRepositoryStub) GetMetrics() (map[string]float64, map[string]int64) {
	return r.gauges, r.counters
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

	value, ok := service.GetGauge("temperature")
	if !ok || value != 12.5 {
		t.Fatalf("expected stored gauge value %v, got %v (exists=%t)", 12.5, value, ok)
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

	gauges, counters := service.GetMetrics()
	if len(gauges) != 0 {
		t.Fatalf("expected no gauges, got %v", gauges)
	}
	if counters["requests"] != 3 {
		t.Fatalf("expected stored counter value %d, got %d", 3, counters["requests"])
	}
}
