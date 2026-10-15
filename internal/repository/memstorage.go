package repository

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (s *MemStorage) AddGauge(name string, value float64) {
	s.Gauges[name] = value
}

func (s *MemStorage) AddCounter(name string, value int64) {
	s.Counters[name] += value
}
