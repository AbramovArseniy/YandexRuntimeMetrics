package server

import (
	"io"
	"net/http"
	"time"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type fileHandler struct {
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

type Server struct {
	Addr        string
	storage     MemStorage
	FileHandler fileHandler
	Debug       bool
	Key         string
}

func NewServer(address string, storeInterval time.Duration, storeFile string, restore bool, debug bool, key string) *Server {
	return &Server{
		Addr: address,
		storage: MemStorage{
			CounterMetrics: make(map[string]int64),
			GaugeMetrics:   make(map[string]float64),
		},
		FileHandler: fileHandler{
			StoreInterval: storeInterval,
			StoreFile:     storeFile,
			Restore:       restore,
		},
		Debug: debug,
		Key:   key,
	}
}
