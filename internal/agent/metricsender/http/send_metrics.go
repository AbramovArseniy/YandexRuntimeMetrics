package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/types"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

type Sender struct {
	client           *http.Client
	UpdateAddress    string
	UpdateAllAddress string
	HostAddress      string
	Key              string
	CryptoKey        *rsa.PublicKey
	RateLimit        int
}

func NewSender(cfg config.Config) *Sender {
	var cryptoKey *rsa.PublicKey
	if cfg.CryptoKeyFile != "" {
		file, err := os.OpenFile(cfg.CryptoKeyFile, os.O_RDONLY, 0777)
		if err != nil {
			loggers.ErrorLogger.Println("error while opening crtypto key file:", err)
			cryptoKey = nil
		} else {
			cryptoKeyByte, err := io.ReadAll(file)
			if err != nil {
				loggers.ErrorLogger.Println("error while opening crtypto key file:", err)
				cryptoKey = nil
			}
			var ok bool
			key, err := x509.ParsePKIXPublicKey(cryptoKeyByte)
			cryptoKey, ok = key.(*rsa.PublicKey)
			if err != nil || !ok {
				loggers.ErrorLogger.Println("error while opening crtypto key file:", err)
				cryptoKey = nil
			}
		}
	}
	return &Sender{
		client:           &http.Client{},
		UpdateAddress:    fmt.Sprintf("http://%s/update/", cfg.Address),
		UpdateAllAddress: fmt.Sprintf("http://%s/updates/", cfg.Address),
		HostAddress:      cfg.HostAddress,
		Key:              cfg.HashKey,
		CryptoKey:        cryptoKey,
		RateLimit:        cfg.RateLimit,
	}
}

// metricWorker gets metrics from channel and sends them to the server
type metricWorker struct {
	ch     chan types.Metrics
	sender *Sender
	mu     sync.Mutex
}

// Compress compresses data sent to the server
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

// SendMetric sends one metric from
func (w *metricWorker) SendMetric() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for metric := range w.ch {
		url := w.sender.UpdateAddress
		if w.sender.Key != "" {
			if metric.MType == "gauge" {
				metric.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), w.sender.Key)
			} else {
				metric.Hash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), w.sender.Key)
			}
		}
		byteJSON, err := json.Marshal(metric)
		if err != nil {
			loggers.ErrorLogger.Println("json Marshal error:", err)
			return err
		}
		if w.sender.CryptoKey != nil {
			codedJSON, err := rsa.EncryptPKCS1v15(rand.Reader, w.sender.CryptoKey, byteJSON)
			if err != nil {
				loggers.ErrorLogger.Println("error while coding json:", err)
			} else {
				byteJSON = codedJSON
			}
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
		req.Header.Set("X-Real-IP", w.sender.HostAddress)
		resp, err := w.sender.client.Do(req)
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

// ReadMetrics sends all metrics to channel
func (w *metricWorker) ReadMetrics(ctx context.Context, collector *metriccollector.MetricCollector) {

	newMetrics := collector.CollectRandomValueMetric()
	metrics := collector.RuntimeMetrics
	metrics = append(metrics, newMetrics)
	metrics = append(metrics, collector.UtilData.CPUutilizations...)
	metrics = append(metrics, collector.UtilData.TotalMemory, collector.UtilData.FreeMemory)
	collector.RuntimeMetrics = append(metrics, newMetrics, collector.PollCount)
	for _, metric := range collector.RuntimeMetrics {
		select {
		case <-ctx.Done():
			return
		case w.ch <- metric:
		}

	}
}

// SendAllMetrics sends all metrics to the server one by one
func (s Sender) SendAllMetrics(collector *metriccollector.MetricCollector) {
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	recordCh := make(chan types.Metrics)
	for i := 0; i < s.RateLimit; i++ {
		w := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: &s}
		g.Go(w.SendMetric)
	}
	readW := &metricWorker{ch: recordCh, mu: sync.Mutex{}, sender: &s}
	readW.ReadMetrics(ctx, collector)
	close(recordCh)
	err := g.Wait()
	if err != nil {
		loggers.ErrorLogger.Println("error sending metrics:", err)
	}
	*(collector.PollCount.Delta) = 0
	loggers.InfoLogger.Println("Sent Gauge")
}

// SendAllMetricsAsButch sends all metrics at one time
func (s Sender) SendAllMetricsAsButch(collector *metriccollector.MetricCollector) {
	var metricHash string
	url := s.UpdateAllAddress
	newMetrics := collector.CollectRandomValueMetric()
	collector.RuntimeMetrics = append(collector.RuntimeMetrics, newMetrics)
	collector.RuntimeMetrics = append(collector.RuntimeMetrics, collector.PollCount)
	var metrics []types.Metrics
	for _, metric := range collector.RuntimeMetrics {
		if s.Key != "" {
			if metric.MType == "gauge" {
				metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
			} else {
				metricHash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), s.Key)
			}
			metric.Hash = metricHash
		}
		metrics = append(metrics, metric)
	}
	for _, metric := range collector.UtilData.CPUutilizations {
		if s.Key != "" {
			metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
			metric.Hash = metricHash
		}
		metrics = append(metrics, metric)
	}
	metric := collector.UtilData.TotalMemory
	if s.Key != "" {
		metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
		metric.Hash = metricHash
	}
	metrics = append(metrics, metric)
	metric = collector.UtilData.FreeMemory
	if s.Key != "" {
		metricHash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
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
	req.Header.Set("X-Real-IP", s.HostAddress)
	resp, err := s.client.Do(req)
	if err != nil {
		loggers.ErrorLogger.Println("Client.Do() error:", err)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		loggers.ErrorLogger.Println("response body close error:", err)
	}
	*(collector.PollCount.Delta) = 0
}
