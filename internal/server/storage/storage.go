package storage

import "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"

// Storage stores metric info
type Storage interface {
	// GetAllMetrics gets info about all metrics
	GetAllMetrics() ([]types.Metrics, error)
	// GetMetrics gets info about one metric
	GetMetric(m types.Metrics, key string) (types.Metrics, error)
	// SaveMetric saves info about one metric
	SaveMetric(metric types.Metrics, key string) error
	// SaveManyMetrics saves info about several metrics
	SaveManyMetrics(metric []types.Metrics, key string) error
	// Check checks if storage works OK
	Check() error
}
