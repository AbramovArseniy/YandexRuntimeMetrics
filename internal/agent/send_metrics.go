package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"golang.org/x/sync/errgroup"
)

type metricWorker struct {
	ch chan Metrics
	a  *Agent
	mu sync.Mutex
}

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

func (w *metricWorker) SendMetric() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for metric := range w.ch {
		url := w.a.UpdateAddress
		if w.a.Key != "" {
			if metric.MType == "gauge" {
				metric.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), w.a.Key)
			} else {
				metric.Hash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), w.a.Key)
			}
		}
		byteJSON, err := json.Marshal(metric)
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
		resp, err := w.a.sender.client.Do(req)
		if err != nil {
			loggers.ErrorLogger.Println("Client.Do() error:", err)
			return err
		}
		if err := resp.Body.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (w *metricWorker) ReadMetrics(ctx context.Context) {
	newMetrics := w.a.collector.CollectRandomValueMetric()
	metrics := w.a.collector.RuntimeMetrics
	metrics = append(metrics, newMetrics)
	metrics = append(metrics, w.a.UtilData.CPUutilizations...)
	metrics = append(metrics, w.a.UtilData.TotalMemory, w.a.UtilData.FreeMemory)
	w.a.collector.RuntimeMetrics = append(metrics, newMetrics, w.a.collector.PollCount)
	for _, metric := range w.a.collector.RuntimeMetrics {
		select {
		case <-ctx.Done():
			return
		case w.ch <- metric:
		}

	}
}

func (a *Agent) SendAllMetrics() {
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	recordCh := make(chan Metrics)
	for i := 0; i < a.RateLimit; i++ {
		w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
		g.Go(w.SendMetric)
	}
	readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, a: a}
	readW.ReadMetrics(ctx)
	close(recordCh)
	err := g.Wait()
	if err != nil {
		loggers.ErrorLogger.Println("error sending metrics:", err)
	}
	*(a.collector.PollCount.Delta) = 0
	loggers.InfoLogger.Println("Sent Gauge")
}

func (a *Agent) SendAllMetricsAsButch() {
	var metricHash string
	url := a.UpdateAllAddress
	newMetrics := a.collector.CollectRandomValueMetric()
	a.collector.RuntimeMetrics = append(a.collector.RuntimeMetrics, newMetrics)
	a.collector.RuntimeMetrics = append(a.collector.RuntimeMetrics, a.collector.PollCount)
	var metrics []Metrics
	for _, metric := range a.collector.RuntimeMetrics {
		if a.Key != "" {
			if metric.MType == "gauge" {
				metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), a.Key)
			} else {
				metricHash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), a.Key)
			}
			metric.Hash = metricHash
		}
		metrics = append(metrics, metric)
	}
	for _, metric := range a.UtilData.CPUutilizations {
		if a.Key != "" {
			metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), a.Key)
			metric.Hash = metricHash
		}
		metrics = append(metrics, metric)
	}
	metric := a.UtilData.TotalMemory
	if a.Key != "" {
		metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), a.Key)
		metric.Hash = metricHash
	}
	metrics = append(metrics, metric)
	metric = a.UtilData.FreeMemory
	if a.Key != "" {
		metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), a.Key)
		metric.Hash = metricHash
	}
	metrics = append(metrics, metric)

	loggers.InfoLogger.Println("Sent Metrics")
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		loggers.ErrorLogger.Println("cannot marshal metrics: " + err.Error())
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
	resp, err := a.sender.client.Do(req)
	if err != nil {
		loggers.ErrorLogger.Println("Client.Do() error:", err)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		loggers.ErrorLogger.Println("response body close error:", err)
	}
	*(a.collector.PollCount.Delta) = 0
}
