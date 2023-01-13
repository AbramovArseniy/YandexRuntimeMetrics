package agent

import (
	"fmt"
	"log"
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

type metricSender struct {
	client *http.Client
}

func NewSender() *metricSender {
	return &metricSender{
		client: &http.Client{},
	}
}

type Agent struct {
	sender    *metricSender
	collector *metricCollector
}

func NewAgent() *Agent {
	return &Agent{
		sender:    NewSender(),
		collector: newCollector(),
	}
}

func (s *metricSender) SendGauge(metric Gauge) error {
	url := fmt.Sprintf("%s%s:%s/update/gauge/%s/%f", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (s *metricSender) SendCounter(metric Counter) error {
	url := fmt.Sprintf("%s%s:%s/update/counter/%s/%d", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (a *Agent) SendAllMetrics() {
	newMetrics := a.collector.CollectRandomValueMetric()
	a.collector.GaugeMetrics = append(a.collector.GaugeMetrics, newMetrics)
	for _, metric := range a.collector.GaugeMetrics {
		err := a.sender.SendGauge(metric)
		if err != nil {
			log.Println("can't send Gauge " + err.Error())
			return
		}
	}
	log.Println("Sent Gauge")
	metricCounter := Counter{metricName: "PollCount", metricValue: a.collector.PollCount}
	a.collector.PollCount = 0
	err := a.sender.SendCounter(metricCounter)
	if err != nil {
		log.Println("can't send Counter " + err.Error())
		return
	}
	log.Println("Sent Counter")
	a.sender.client.CloseIdleConnections()
}
