package storage

import "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"

type Storage interface {
	GetAllMetrics() ([]types.Metrics, error)
	GetMetric(m types.Metrics, key string) (types.Metrics, error)
	SaveMetric(metric types.Metrics, key string) error
	SaveManyMetrics(metric []types.Metrics, key string) error
	Check() error
}

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}
