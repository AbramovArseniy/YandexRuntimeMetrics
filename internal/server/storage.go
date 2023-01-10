package server

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}
