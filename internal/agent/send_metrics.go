package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
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

func hash(src, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}

func (a *Agent) SendGauge(metric Gauge) error {
	url := a.UpdateAddress
	m := Metrics{
		ID:    metric.metricName,
		MType: "gauge",
		Value: &metric.metricValue,
	}
	if a.Key != "" {
		m.Hash = hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), a.Key)
	}
	byteJSON, err := json.Marshal(m)
	if err != nil {
		loggers.ErrorLogger.Println("json Marshal error:", err)
		return err
	}
	compressedJSON, err := Compress(byteJSON)
	if err != nil {
		loggers.ErrorLogger.Printf("Compress error: %v", err)
	}
	body := strings.NewReader(string(compressedJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		loggers.ErrorLogger.Println("Request Creation error:", err)
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := a.sender.client.Do(req)
	if err != nil {
		loggers.ErrorLogger.Println("Client.Do() error:", err)
		return err
	}
	return resp.Body.Close()
}

func (a *Agent) SendCounter(metric Counter) error {
	url := a.UpdateAddress
	m := Metrics{
		ID:    metric.metricName,
		MType: "counter",
		Delta: &metric.metricValue,
	}
	if a.Key != "" {
		m.Hash = hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), a.Key)
	}
	byteJSON, err := json.Marshal(m)
	if err != nil {
		loggers.ErrorLogger.Println("json Marshal error:", err)
		return err
	}
	compressedJSON, err := Compress(byteJSON)
	if err != nil {
		loggers.ErrorLogger.Printf("Compress error: %v", err)
	}
	body := strings.NewReader(string(compressedJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		loggers.ErrorLogger.Println("Request Creation error")
		return err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := a.sender.client.Do(req)
	if err != nil {
		loggers.ErrorLogger.Println("Client.Do() error:", err)
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
			loggers.ErrorLogger.Println("can't send Gauge " + err.Error())
			return
		}
	}
	loggers.InfoLogger.Println("Sent Gauge")
	metricCounter := Counter{metricName: "PollCount", metricValue: a.collector.PollCount}
	a.collector.PollCount = 0
	err := a.SendCounter(metricCounter)
	if err != nil {
		loggers.ErrorLogger.Println("can't send Counter " + err.Error())
		return
	}
	loggers.InfoLogger.Println("Sent Counter")
}

func (a *Agent) SendAllMetricsAsButch() {
	url := a.UpdateAllAddress
	newMetrics := a.collector.CollectRandomValueMetric()
	a.collector.GaugeMetrics = append(a.collector.GaugeMetrics, newMetrics)
	var metrics []Metrics
	for _, metric := range a.collector.GaugeMetrics {
		var metricHash string
		if a.Key != "" {
			metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.metricName, metric.metricValue), a.Key)
		}
		var value = metric.metricValue
		metrics = append(metrics, Metrics{
			ID:    metric.metricName,
			MType: "gauge",
			Value: &value,
			Hash:  metricHash,
		})
		loggers.DebugLogger.Println(metrics, metric.metricName, metric.metricValue)
	}
	loggers.InfoLogger.Println("Sent Gauge")
	var metricHash string
	if a.Key != "" {
		metricHash = hash(fmt.Sprintf("%s:counter:%d", "PollCount", a.collector.PollCount), a.Key)
	}
	var delta = a.collector.PollCount
	metrics = append(metrics, Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &delta,
		Hash:  metricHash,
	})
	a.collector.PollCount = 0
	jsonMetrics, err := json.Marshal(metrics)
	loggers.DebugLogger.Println(string(jsonMetrics))
	if err != nil {
		loggers.ErrorLogger.Println("can't send Counter " + err.Error())
		return
	}
	compressedJSON, err := Compress(jsonMetrics)
	if err != nil {
		loggers.ErrorLogger.Printf("Compress error: %v", err)
	}
	body := strings.NewReader(string(compressedJSON))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		loggers.ErrorLogger.Println("Request Creation error")
		return
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	_, err = a.sender.client.Do(req)
	if err != nil {
		loggers.ErrorLogger.Println("Client.Do() error:", err)
		return
	}
	loggers.InfoLogger.Println("Sent Counter")
}
