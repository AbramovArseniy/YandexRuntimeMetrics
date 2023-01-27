package server

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

func (s *Server) StoreMetricsToFile() {
	file, err := os.OpenFile(s.FileHandler.StoreFile, os.O_WRONLY|os.O_CREATE, 0777)
	writer := bufio.NewWriter(file)
	if err != nil {
		loggers.ErrorLogger.Printf("Failed to open file: %s", s.FileHandler.StoreFile)
	}
	defer file.Close()
	for name, value := range s.storage.CounterMetrics {
		gauge := Metrics{
			ID:    name,
			Delta: &value,
			MType: "counter",
		}
		jsonMetric, err := json.Marshal(gauge)
		if err != nil {
			loggers.ErrorLogger.Printf("error marshaling json to file: %v", err)
			return
		}
		_, err = writer.Write(jsonMetric)
		if err != nil {
			loggers.ErrorLogger.Printf("error writing to file: %v", err)
			return
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			loggers.ErrorLogger.Printf("error writing to file: %v", err)
			return
		}
	}
	for name, value := range s.storage.GaugeMetrics {
		gauge := Metrics{
			ID:    name,
			Value: &value,
			MType: "gauge",
		}
		jsonMetric, err := json.Marshal(gauge)
		if err != nil {
			loggers.ErrorLogger.Printf("error marshaling json to file: %v", err)
			return
		}
		_, err = writer.Write(jsonMetric)
		if err != nil {
			loggers.ErrorLogger.Printf("error writing to file: %v", err)
			return
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			loggers.ErrorLogger.Printf("error writing to file: %v", err)
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		loggers.ErrorLogger.Printf("writer.Flush() error: %v", err)
	}
	loggers.InfoLogger.Println("stored to file")
}

func (s *Server) RestoreMetricsFromFile() {
	file, err := os.OpenFile(s.FileHandler.StoreFile, os.O_RDONLY|os.O_CREATE, 0777)
	scanner := bufio.NewScanner(file)
	if err != nil {
		loggers.ErrorLogger.Printf("Failed to open file: %s, %v", s.FileHandler.StoreFile, err)
	}
	defer file.Close()
	for scanner.Scan() {
		m := Metrics{}
		err = json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			loggers.ErrorLogger.Printf("json Unmarshal error: %v", err)
			return
		}
		s.storeMetrics(m)
	}
	loggers.InfoLogger.Printf("Restored Metrics from '%s'", s.FileHandler.StoreFile)
}
