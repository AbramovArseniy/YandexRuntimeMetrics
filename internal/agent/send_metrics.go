package agent

import (
	"fmt"
	"net/http"
	"strings"
)

type Gauge struct {
	metricName  string
	metricValue float64
}

type Counter struct {
	metricName  string
	metricValue int64
}

func (metric Gauge) SendGauge(client *http.Client) (bool, error) {
	url := fmt.Sprintf("%s%s:%s/update/gauge/%s/%f", Protocol, Server, Port, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := client.Post(url, "text/plain", body)
	if err != nil {
		return false, err
	}
	err = resp.Body.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (metric Counter) SendCounter(client *http.Client) (bool, error) {
	url := fmt.Sprintf("%s%s:%s/update/counter/%s/%d", Protocol, Server, Port, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := client.Post(url, "text/plain", body)
	if err != nil {
		return false, err
	}
	err = resp.Body.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}

func SendAllMetrics() {
	Metrics = CollectRandomValueMetric(Metrics)
	client := http.Client{Timeout: Timeout}
	for _, i := range Metrics {
		_, err := i.SendGauge(&client)
		if err != nil {
			return
		}
	}
	metricCounter := Counter{metricName: "PollCount", metricValue: PollCount}
	PollCount = 0
	_, err := metricCounter.SendCounter(&client)
	if err != nil {
		return
	}
	client.CloseIdleConnections()
}
