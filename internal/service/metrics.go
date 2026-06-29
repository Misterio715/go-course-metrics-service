package service

type MetricsRepository interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
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
