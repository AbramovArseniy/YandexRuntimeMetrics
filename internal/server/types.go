package server

import (
	"fmt"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

type fileHandler struct {
	StoreInterval int
	StoreFile     string
	Restore       bool
}

type Server struct {
	Addr        string
	handler     Handler
	FileHandler fileHandler
}

func NewServer() *Server {
	return &Server{
		Addr: "localhost:8080",
		handler: Handler{
			storage: MemStorage{
				CounterMetrics: make(map[string]int64),
				GaugeMetrics:   make(map[string]float64),
			},
		},
		FileHandler: fileHandler{
			StoreInterval: 300,
			StoreFile:     "/tmp/devops-metrics-db.json",
			Restore:       true,
		},
	}
}

func (s Server) String() string {
	return fmt.Sprintf("Storage: %v \file Storing:%v", s.handler.storage, s.FileHandler)
}
