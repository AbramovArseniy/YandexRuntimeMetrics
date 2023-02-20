package agent

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type UtilizationData struct {
	mu              sync.Mutex
	TotalMemory     Metrics
	FreeMemory      Metrics
	CPUutilizations []Metrics
	CPUtime         []float64
	CPUutilLastTime time.Time
}

type metricCollector struct {
	RuntimeMetrics []Metrics
	PollCount      Metrics
}

func newCollector() *metricCollector {
	var delta int64 = 0
	return &metricCollector{
		PollCount: Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		},
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
	sender           *metricSender
	collector        *metricCollector
	Address          string
	UpdateAddress    string
	UpdateAllAddress string
	PollInterval     time.Duration
	Key              string
	ReportInterval   time.Duration
	RateLimit        int
	UtilData         UtilizationData
}

func NewAgent(addr string, pollInterval time.Duration, reportInterval time.Duration, key string, rateLimit int) *Agent {
	return &Agent{
		Address:          addr,
		UpdateAddress:    fmt.Sprintf("http://%s/update/", addr),
		UpdateAllAddress: fmt.Sprintf("http://%s/updates/", addr),
		sender:           NewSender(),
		collector:        newCollector(),
		PollInterval:     pollInterval,
		ReportInterval:   reportInterval,
		Key:              key,
		RateLimit:        rateLimit,
	}
}
