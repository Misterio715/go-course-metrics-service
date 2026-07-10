package repository

import "sync"

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) AddGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

func (s *MemStorage) AddCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.gauges[name]
	return value, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.counters[name]
	return value, ok
}

func (s *MemStorage) GetMetrics() (map[string]float64, map[string]int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gauges := make(map[string]float64, len(s.gauges))
	for name, value := range s.gauges {
		gauges[name] = value
	}

	counters := make(map[string]int64, len(s.counters))
	for name, value := range s.counters {
		counters[name] = value
	}

	return gauges, counters
}
