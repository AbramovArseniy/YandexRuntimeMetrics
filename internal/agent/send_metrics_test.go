package agent

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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
		want        bool
		expectError bool
	}{
		{
			name: "OK Result",
			metric: metric{
				name:  "PollCount",
				value: 5,
			},
			client:      &http.Client{Timeout: Timeout},
			want:        true,
			expectError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Counter{
				metricName:  test.metric.name,
				metricValue: test.metric.value,
			}

			l, err := net.Listen(TCP, Server+":"+Port)
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

			got, err := metric.SendCounter(test.client)
			if (err != nil) != test.expectError {
				t.Errorf("counter.SendCounter() error = %v, expectError %v", err, test.expectError)
				return
			}
			if got != test.want {
				t.Errorf("counter.SendCounter() = %v, want %v", got, test.want)
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
		want        bool
		expectError bool
	}{
		{
			name: "OK Result",
			metric: metric{
				name:  "PollCount",
				value: 5.5,
			},
			client:      &http.Client{Timeout: Timeout},
			want:        true,
			expectError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := Gauge{
				metricName:  test.metric.name,
				metricValue: test.metric.value,
			}

			l, err := net.Listen(TCP, Server+":"+Port)
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

			got, err := metric.SendGauge(test.client)
			if (err != nil) != test.expectError {
				t.Errorf("counter.SendGauge() error = %v, expectError %v", err, test.expectError)
				return
			}
			if got != test.want {
				t.Errorf("counter.SendGauge() = %v, want %v", got, test.want)
			}
		})
	}

}
