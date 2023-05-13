package agent

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// Metrics saves metric info
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// UtilizationData collects cpu and mem metrics
type UtilizationData struct {
	mu              sync.Mutex
	TotalMemory     Metrics
	FreeMemory      Metrics
	CPUutilizations []Metrics
	CPUtime         []float64
	CPUutilLastTime time.Time
}

// metricCollector collects metrics
type metricCollector struct {
	RuntimeMetrics []Metrics
	PollCount      Metrics
}

// newCollector creates a new metricCollector
func newCollector() *metricCollector {
	var delta int64 = 0
	return &metricCollector{
		PollCount: Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		},
	}
}

// metricSender sends all metrics to the server
type metricSender struct {
	client *http.Client
}

// NewSender created new metricSender
func NewSender() *metricSender {
	return &metricSender{
		client: &http.Client{},
	}
}

// Agent makes all the work with metrics
type Agent struct {
	sender           *metricSender
	collector        *metricCollector
	Address          string
	UpdateAddress    string
	UpdateAllAddress string
	PollInterval     time.Duration
	Key              string
	ReportInterval   time.Duration
	RateLimit        int
	UtilData         UtilizationData
	CryptoKey        *rsa.PublicKey
	HostAddress      string
}

// NewAgent creates new Agent
func NewAgent(cfg config.Config) *Agent {
	conn, err := net.Dial("tcp", cfg.Address)
	if err != nil {
		loggers.ErrorLogger.Println("error while making connection:", err)
	}
	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		loggers.ErrorLogger.Println("error while splitting host and port:", err)
	}
	conn.Close()
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
	return &Agent{
		Address:          cfg.Address,
		UpdateAddress:    fmt.Sprintf("http://%s/update/", cfg.Address),
		UpdateAllAddress: fmt.Sprintf("http://%s/updates/", cfg.Address),
		sender:           NewSender(),
		collector:        newCollector(),
		PollInterval:     cfg.PollInterval,
		ReportInterval:   cfg.ReportInterval,
		Key:              cfg.HashKey,
		RateLimit:        cfg.RateLimit,
		CryptoKey:        cryptoKey,
		HostAddress:      host,
	}
}
