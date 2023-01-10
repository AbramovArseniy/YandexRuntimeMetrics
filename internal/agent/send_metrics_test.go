package agent

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	TCP            = "tcp"
	DefaultTimeout = 2 * time.Second
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
			client:      &http.Client{Timeout: DefaultTimeout},
			expectError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Counter{
				metricName:  test.metric.name,
				metricValue: test.metric.value,
			}
			s := NewSender()
			l, err := net.Listen(TCP, DefaultHost+":"+DefaultPort)
			if err != nil {
				log.Fatal(err)
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

			err = s.SendCounter(metric)
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
				value: 5.5,
			},
			client:      &http.Client{Timeout: DefaultTimeout},
			expectError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Gauge{
				metricName:  test.metric.name,
				metricValue: test.metric.value,
			}
			s := NewSender()
			l, err := net.Listen(TCP, DefaultHost+":"+DefaultPort)
			if err != nil {
				log.Fatal(err)
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

			err = s.SendGauge(metric)
			if (err != nil) != test.expectError {
				t.Errorf("counter.SendGauge() error = %v, expectError %v", err, test.expectError)
				return
			}
		})
	}

}
