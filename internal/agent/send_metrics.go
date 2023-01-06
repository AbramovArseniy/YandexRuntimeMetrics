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

func (s *metricSender) SendGauge(metric Gauge) (bool, error) {
	url := fmt.Sprintf("%s%s:%s/update/gauge/%s/%f", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}

func (s *metricSender) SendCounter(metric Counter) (bool, error) {
	url := fmt.Sprintf("%s%s:%s/update/gauge/%s/%d", DefaultProtocol, DefaultHost, DefaultPort, metric.metricName, metric.metricValue)
	body := strings.NewReader("")
	resp, err := s.client.Post(url, "text/plain", body)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	return true, nil
}

func SendAllMetrics() {
	s := NewSender()
	newMetrics := CollectRandomValueMetric()
	allMetrics = append(allMetrics, newMetrics)
	for _, metric := range allMetrics {
		_, err := s.SendGauge(metric)
		if err != nil {
			log.Println("can't send Gauge " + err.Error())
			return
		}
	}
	metricCounter := Counter{metricName: "PollCount", metricValue: PollCount}
	PollCount = 0
	_, err := s.SendCounter(metricCounter)
	if err != nil {
		log.Println("can't send Counter " + err.Error())
		return
	}
	s.client.CloseIdleConnections()
}
