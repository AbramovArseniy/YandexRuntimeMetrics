package agent

import (
	"fmt"
	"net"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metricsender"
	grpc "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metricsender/grpc"
	http "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metricsender/http"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// Agent makes all the work with metrics
type Agent struct {
	Sender         metricsender.MetricSender
	Collector      *metriccollector.MetricCollector
	PollInterval   time.Duration
	ReportInterval time.Duration
}

// NewAgent creates new Agent
func NewAgent(cfg config.Config) (*Agent, error) {
	var host string
	conn, err := net.Dial("tcp", cfg.Address)
	if err != nil {
		loggers.ErrorLogger.Println("error while making connection:", err)
	} else {
		host, _, err = net.SplitHostPort(conn.LocalAddr().String())
		if err != nil {
			loggers.ErrorLogger.Println("error while splitting host and port:", err)
		}
	}
	cfg.HostAddress = host
	conn.Close()
	var sender metricsender.MetricSender
	if cfg.Protocol == "HTTP" {
		sender = *http.NewSender(cfg)
	} else if cfg.Protocol == "gRPC" {
		sender = *grpc.NewSender(cfg)
	} else {
		return nil, fmt.Errorf("wrong protocol")
	}
	return &Agent{
		Sender:         sender,
		Collector:      metriccollector.NewMetricCollector(),
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
	}, nil
}

func (a *Agent) SendAllMetrics() {
	a.Sender.SendAllMetrics(a.Collector)
}

func (a *Agent) SendAllMetricsAsButch() {
	a.Sender.SendAllMetricsAsButch(a.Collector)
}
