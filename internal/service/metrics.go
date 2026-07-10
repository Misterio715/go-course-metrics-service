package service

type MetricsRepository interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetMetrics() (map[string]float64, map[string]int64)
}

type MetricsService struct {
	repository MetricsRepository
}

func NewMetricsService(repository MetricsRepository) *MetricsService {
	return &MetricsService{repository: repository}
}

func (s *MetricsService) AddGauge(name string, value float64) {
	s.repository.AddGauge(name, value)
}

func (s *MetricsService) AddCounter(name string, value int64) {
	s.repository.AddCounter(name, value)
}

func (s *MetricsService) GetGauge(name string) (float64, bool) {
	return s.repository.GetGauge(name)
}

func (s *MetricsService) GetCounter(name string) (int64, bool) {
	return s.repository.GetCounter(name)
}

func (s *MetricsService) GetMetrics() (map[string]float64, map[string]int64) {
	return s.repository.GetMetrics()
}
