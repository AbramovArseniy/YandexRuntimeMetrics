package server

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

var Storage = MemStorage{
	CounterMetrics: make(map[string]int64),
	GaugeMetrics:   make(map[string]float64),
}
