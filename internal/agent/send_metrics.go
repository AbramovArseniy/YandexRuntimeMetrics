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

type agent struct {
	sender *metricSender
}

func NewAgent() *agent {
	return &agent{
		sender: NewSender(),
	}
}

var a = NewAgent()

func (s *metricSender) SendGauge(metric Gauge) error {
	url := fmt.Sprintf("%s%s:%s/update/gauge/%s/%f", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *metricSender) SendCounter(metric Counter) error {
	url := fmt.Sprintf("%s%s:%s/update/counter/%s/%d", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func SendAllMetrics() {
	newMetrics := CollectRandomValueMetric()
	allMetrics = append(allMetrics, newMetrics)
	for _, metric := range allMetrics {
		err := a.sender.SendGauge(metric)
		if err != nil {
			log.Println("can't send Gauge " + err.Error())
			return
		}
	}
	log.Println("Sent Gauge")
	metricCounter := Counter{metricName: "PollCount", metricValue: PollCount}
	PollCount = 0
	err := a.sender.SendCounter(metricCounter)
	if err != nil {
		log.Println("can't send Counter " + err.Error())
		return
	}
	log.Println("Sent Counter")
	a.sender.client.CloseIdleConnections()
}
