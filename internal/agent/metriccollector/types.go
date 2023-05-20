package metriccollector

import (
	"sync"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/types"
)

// UtilizationData collects cpu and mem metrics
type UtilizationData struct {
	mu              sync.Mutex
	TotalMemory     types.Metrics
	FreeMemory      types.Metrics
	CPUutilizations []types.Metrics
	CPUtime         []float64
	CPUutilLastTime time.Time
}

// metricCollector collects metrics
type MetricCollector struct {
	UtilData       UtilizationData
	RuntimeMetrics []types.Metrics
	PollCount      types.Metrics
}

// newCollector creates a new MetricCollector
func NewMetricCollector() *MetricCollector {
	var delta int64 = 0
	return &MetricCollector{
		UtilData: UtilizationData{
			mu: sync.Mutex{},
		},
		PollCount: types.Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		},
	}
}
