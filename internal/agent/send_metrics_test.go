package agent

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"golang.org/x/sync/errgroup"
)

// Default test preferences
const (
	tcp            = "tcp"
	defaultTimeout = 2 * time.Second
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
		Address:        "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      100,
	}
	a := NewAgent(cfg)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Metrics{
				ID:    test.metric.name,
				MType: "counter",
				Delta: &test.metric.value,
			}
			l, err := net.Listen(tcp, DefaultHost+":"+DefaultPort)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			serveMux := http.NewServeMux()
			serveMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewUnstartedServer(serveMux)
			err = srv.Listener.Close()
			if err != nil {
				return
			}
			srv.Listener = l
			srv.Start()

			defer srv.Close()
			ctx := context.Background()
			g, _ := errgroup.WithContext(ctx)
			recordCh := make(chan Metrics)
			a.collector.RuntimeMetrics = []Metrics{metric}
			for i := 0; i < a.RateLimit; i++ {
				w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
				g.Go(w.SendMetric)
			}
			readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
			readW.ReadMetrics(ctx)
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
		Address:        "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      100,
	}
	a := NewAgent(cfg)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Metrics{
				ID:    test.metric.name,
				Value: &test.metric.value,
			}
			l, err := net.Listen(tcp, DefaultHost+":"+DefaultPort)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			serveMux := http.NewServeMux()
			serveMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewUnstartedServer(serveMux)
			err = srv.Listener.Close()
			if err != nil {
				return
			}
			srv.Listener = l
			srv.Start()

			defer srv.Close()

			ctx := context.Background()
			g, _ := errgroup.WithContext(ctx)
			recordCh := make(chan Metrics)
			a.collector.RuntimeMetrics = []Metrics{metric}
			for i := 0; i < a.RateLimit; i++ {
				w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
				g.Go(w.SendMetric)
			}
			readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
			readW.ReadMetrics(ctx)
			close(recordCh)
			err = g.Wait()
			if (err != nil) != test.expectError {
				t.Errorf("error Sending Gauge error = %v, expectError %v", err, test.expectError)
				return
			}
		})
	}

}
