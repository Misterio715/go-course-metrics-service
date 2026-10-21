package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	models "github.com/Misterio715/go-course-metrics-service/internal/model"
	"github.com/rs/zerolog/log"
)

type FileStorage struct {
	*MemStorage
	filePath string
	syncSave bool
	saveMu   sync.Mutex
}

func NewFileStorage(filePath string, syncSave bool) *FileStorage {
	return &FileStorage{
		MemStorage: NewMemStorage(),
		filePath:   filePath,
		syncSave:   syncSave,
	}
}

func (s *FileStorage) AddGauge(name string, value float64) {
	s.MemStorage.AddGauge(name, value)
	s.saveAfterUpdate()
}

func (s *FileStorage) AddCounter(name string, value int64) {
	s.MemStorage.AddCounter(name, value)
	s.saveAfterUpdate()
}

func (s *FileStorage) saveAfterUpdate() {
	if s.syncSave {
		if err := s.Save(); err != nil {
			log.Error().Err(err).Str("path", s.filePath).Msg("Failed to save metrics")
		}
	}
}

func (s *FileStorage) Save() error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	gauges, counters := s.GetMetrics()
	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))
	for name, value := range gauges {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Gauge, Value: &value})
	}
	for name, value := range counters {
		metrics = append(metrics, models.Metrics{ID: name, MType: models.Counter, Delta: &value})
	}

	file, err := os.CreateTemp(filepath.Dir(s.filePath), ".metrics-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), s.filePath)
}

func (s *FileStorage) Restore() error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	for _, metric := range metrics {
		if metric.ID == "" {
			return errors.New("stored metric has an empty ID")
		}
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("stored gauge %q has no value", metric.ID)
			}
			gauges[metric.ID] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("stored counter %q has no delta", metric.ID)
			}
			counters[metric.ID] = *metric.Delta
		default:
			return fmt.Errorf("unknown stored metric type %q", metric.MType)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges = gauges
	s.counters = counters
	return nil
}
