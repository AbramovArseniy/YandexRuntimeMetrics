package agent

import (
	//"fmt"
	//"log"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	//"strings"
)

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
	Address   string
}

func NewAgent() *Agent {
	return &Agent{
		Address:   "localhost:8080",
		sender:    NewSender(),
		collector: newCollector(),
	}
}

func (a *Agent) SendGauge(metric Gauge) error {
	url := fmt.Sprintf("http://%s/update/", a.Address)
	m := Metrics{
		ID:    metric.metricName,
		MType: "gauge",
		Value: &metric.metricValue,
	}
	byteJSON, err := json.Marshal(m)
	if err != nil {
		log.Println("json Marshal error")
		return err
	}
	body := strings.NewReader(string(byteJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.sender.client.Do(req)

	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (a *Agent) SendCounter(metric Counter) error {
	url := fmt.Sprintf("http://%s/update/", a.Address)
	m := Metrics{
		ID:    metric.metricName,
		MType: "counter",
		Delta: &metric.metricValue,
	}
	byteJSON, err := json.Marshal(m)
	if err != nil {
		log.Println("json Marshal error")
		return err
	}
	body := strings.NewReader(string(byteJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.sender.client.Do(req)
	if err != nil {
		log.Println(body)
		return err
	}
	return resp.Body.Close()
}

func (a *Agent) SendAllMetrics() {
	newMetrics := a.collector.CollectRandomValueMetric()
	a.collector.GaugeMetrics = append(a.collector.GaugeMetrics, newMetrics)
	for _, metric := range a.collector.GaugeMetrics {
		err := a.SendGauge(metric)
		if err != nil {
			log.Println("can't send Gauge " + err.Error())
			return
		}
	}
	log.Println("Sent Gauge")
	metricCounter := Counter{metricName: "PollCount", metricValue: a.collector.PollCount}
	a.collector.PollCount = 0
	err := a.SendCounter(metricCounter)
	if err != nil {
		log.Println("can't send Counter " + err.Error())
		return
	}
	log.Println("Sent Counter")
}
