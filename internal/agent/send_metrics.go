package agent

import (
	//"fmt"
	//"log"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	//"strings"
	"compress/gzip"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return b.Bytes(), nil
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
	compressedJSON, err := Compress(byteJSON)
	if err != nil {
		log.Printf("Compress error: %v", err)
	}
	body := strings.NewReader(string(compressedJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
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
	compressedJSON, err := Compress(byteJSON)
	if err != nil {
		log.Printf("Compress error: %v", err)
	}
	body := strings.NewReader(string(compressedJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
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
