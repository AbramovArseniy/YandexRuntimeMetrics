package agent

import "net/http"

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Gauge struct {
	metricName  string
	metricValue float64
}

type Counter struct {
	metricName  string
	metricValue int64
}

type metricCollector struct {
	GaugeMetrics []Gauge
	PollCount    int64
}

func newCollector() *metricCollector {
	return &metricCollector{
		GaugeMetrics: make([]Gauge, 0),
		PollCount:    0,
	}
}

type metricSender struct {
	client *http.Client
}

func NewSender() *metricSender {
	return &metricSender{
		client: &http.Client{},
	}
}

type Agent struct {
	sender         *metricSender
	collector      *metricCollector
	Address        string
	PollInterval   int
	ReportInterval int
}

func NewAgent() *Agent {
	return &Agent{
		Address:        "localhost:8080",
		sender:         NewSender(),
		collector:      newCollector(),
		PollInterval:   2,
		ReportInterval: 10,
	}
}
