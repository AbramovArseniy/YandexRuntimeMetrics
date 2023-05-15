// Package grpc describes sending metrics' info with gRPC protocol
package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/proto"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/types"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

type Sender struct {
	Client      pb.MetricsClient
	HostAddress string
	Key         string
	RateLimit   int
}

func NewSender(cfg config.Config) *Sender {
	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		loggers.ErrorLogger.Println("error while making connection:", err)
	}
	client := pb.NewMetricsClient(conn)
	conn.Close()
	return &Sender{
		Client:      client,
		HostAddress: cfg.HostAddress,
		Key:         cfg.HashKey,
		RateLimit:   cfg.RateLimit,
	}
}

// metricWorker gets metrics from channel and sends them to the server
type metricWorker struct {
	ch     chan types.Metrics
	sender *Sender
	mu     sync.Mutex
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
		m := pb.Metric{
			Id:    metric.ID,
			Mtype: metric.MType,
		}
		if w.sender.Key != "" {
			if metric.MType == "gauge" {
				m.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), w.sender.Key)
			} else {
				m.Hash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), w.sender.Key)
			}
		}
		switch metric.MType {
		case "counter":
			m.Delta = *metric.Delta
		case "gauge":
			m.Value = *metric.Value
		}
		req := pb.UpdateMetricRequest{
			Metric: &m,
		}
		_, err := w.sender.Client.UpdateMetric(context.Background(), &req)
		if err != nil {
			if e, ok := status.FromError(err); ok {
				if e.Code() == codes.PermissionDenied {
					loggers.ErrorLogger.Println(`FORBIDDEN`, e.Message())
				} else if e.Code() == codes.Unimplemented {
					loggers.ErrorLogger.Println("UNIMPLEMENTED", e.Message())
				}
				return err
			}
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
	newMetrics := collector.CollectRandomValueMetric()
	collector.RuntimeMetrics = append(collector.RuntimeMetrics, newMetrics)
	collector.RuntimeMetrics = append(collector.RuntimeMetrics, collector.PollCount)
	var metrics []*pb.Metric
	for _, metric := range collector.RuntimeMetrics {
		m := pb.Metric{
			Id:    metric.ID,
			Mtype: metric.MType,
		}
		if s.Key != "" {
			if metric.MType == "gauge" {
				m.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
			} else {
				m.Hash = hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), s.Key)
			}
		}
		switch metric.MType {
		case "counter":
			m.Delta = *metric.Delta
		case "gauge":
			m.Value = *metric.Value
		}
		metrics = append(metrics, &m)
	}
	for _, metric := range collector.UtilData.CPUutilizations {
		m := pb.Metric{
			Id:    metric.ID,
			Mtype: metric.MType,
			Value: *metric.Value,
		}
		if s.Key != "" {
			m.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
		}
		metrics = append(metrics, &m)
	}
	metric := collector.UtilData.TotalMemory
	m := pb.Metric{
		Id:    metric.ID,
		Mtype: metric.MType,
		Value: *metric.Value,
	}
	if s.Key != "" {
		m.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
	}
	metrics = append(metrics, &m)
	metric = collector.UtilData.FreeMemory
	m = pb.Metric{
		Id:    metric.ID,
		Mtype: metric.MType,
		Value: *metric.Value,
	}
	if s.Key != "" {
		m.Hash = hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), s.Key)
	}
	metrics = append(metrics, &m)
	loggers.InfoLogger.Println("Sent Metrics")
	mdMap := make(map[string]string)
	mdMap["X-Real-IP"] = s.HostAddress
	md := metadata.New(mdMap)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	req := &pb.UpdateManyMetricsRequest{
		Metrics: metrics,
	}
	_, err := s.Client.UpdateManyMetrics(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.PermissionDenied {
				loggers.ErrorLogger.Println(`FORBIDDEN`, e.Message())
			} else if e.Code() == codes.Unimplemented {
				loggers.ErrorLogger.Println("UNIMPLEMENTED", e.Message())
			}
			return
		}
	}
	*(collector.PollCount.Delta) = 0
}
