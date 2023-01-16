package agent

import (
	//"fmt"
	//"log"
	"fmt"
	"log"
	"net/http"
	"strings"
	//"strings"
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
	url := fmt.Sprintf("%s%s:%s/update", DefaultProtocol, DefaultHost, DefaultPort)
	body := strings.NewReader(fmt.Sprintf(`{"id":"%s","type":"%s","value":%f}`, metric.metricName, "gauge", metric.metricValue))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)

	if err != nil {
		log.Println("client.Do error")
		return err
	}
	return resp.Body.Close()
}

func (s *metricSender) SendCounter(metric Counter) error {
	url := fmt.Sprintf("%s%s:%s/update", DefaultProtocol, DefaultHost, DefaultPort)
	body := strings.NewReader(fmt.Sprintf(`{"id":"%s","type":"%s","delta":%d}`, metric.metricName, "counter", metric.metricValue))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Println("client.Do error")
		return err
	}
	return resp.Body.Close()
}

func (a *Agent) SendAllMetrics() {
	newMetrics := a.collector.CollectRandomValueMetric()
	log.Println(newMetrics)
	/*newMetrics := a.collector.CollectRandomValueMetric()
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
	log.Println("Sent Counter")*/
}
