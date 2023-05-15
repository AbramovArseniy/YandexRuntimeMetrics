package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/types"
)

// Default test preferences
const (
	tcp            = "tcp"
	defaultTimeout = 2 * time.Second
	defaultHost    = "localhost"
	defaultPort    = "8080"
)

func TestSendCounter(t *testing.T) {
	type metric struct {
		name  string
		value int64
	}
	tests := []struct {
		name        string
		metric      metric
		client      *http.Client
		expectError bool
	}{
		{
			name: "OK Result",
			metric: metric{
				name:  "PollCount",
				value: 5,
			},
			client:      &http.Client{Timeout: defaultTimeout},
			expectError: false,
		},
	}
	cfg := config.Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Address:        "localhost:8080",
		RateLimit:      100,
	}
	c := metriccollector.NewMetricCollector()
	s := NewSender(cfg)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := types.Metrics{
				ID:    test.metric.name,
				MType: "counter",
				Delta: &test.metric.value,
			}
			l, err := net.Listen(tcp, defaultHost+":"+defaultPort)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			serveMux := http.NewServeMux()
			serveMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewUnstartedServer(serveMux)
			srv.Listener.Close()
			srv.Listener = l
			srv.Start()
			defer srv.Close()
			ctx := context.Background()
			g, _ := errgroup.WithContext(ctx)
			recordCh := make(chan types.Metrics)
			c.RuntimeMetrics = []types.Metrics{metric}
			for i := 0; i < s.RateLimit; i++ {
				w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: s}
				g.Go(w.SendMetric)
			}
			readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: s}
			readW.ReadMetrics(ctx, c)
			close(recordCh)
			err = g.Wait()
			if (err != nil) != test.expectError {
				t.Errorf("counter.SendCounter() error = %v, expectError %v", err, test.expectError)
				return
			}
		})
	}

}

func TestSendGauge(t *testing.T) {
	type metric struct {
		name  string
		value float64
	}
	tests := []struct {
		name        string
		metric      metric
		client      *http.Client
		expectError bool
	}{
		{
			name: "OK Result",
			metric: metric{
				name:  "PollCount",
				value: 5.567675674566564534343565445434345343768767867676786786766867676767676767667676767676677,
			},
			client:      &http.Client{Timeout: defaultTimeout},
			expectError: false,
		},
	}
	cfg := config.Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Address:        "localhost:8080",
		RateLimit:      100,
		HostAddress:    "127.0.0.1",
	}
	c := metriccollector.NewMetricCollector()
	s := NewSender(cfg)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := types.Metrics{
				ID:    test.metric.name,
				Value: &test.metric.value,
			}
			l, err := net.Listen(tcp, defaultHost+":"+defaultPort)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			serveMux := http.NewServeMux()
			serveMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewUnstartedServer(serveMux)
			srv.Listener.Close()
			srv.Listener = l
			srv.Start()
			defer srv.Close()

			ctx := context.Background()
			g, _ := errgroup.WithContext(ctx)
			recordCh := make(chan types.Metrics)
			c.RuntimeMetrics = []types.Metrics{metric}
			for i := 0; i < s.RateLimit; i++ {
				w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: s}
				g.Go(w.SendMetric)
			}
			readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: s}
			readW.ReadMetrics(ctx, c)
			close(recordCh)
			err = g.Wait()
			if (err != nil) != test.expectError {
				t.Errorf("error Sending Gauge error = %v, expectError %v", err, test.expectError)
				return
			}
		})
	}

}
