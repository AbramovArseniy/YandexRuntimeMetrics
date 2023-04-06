package filestorage

import (
	"bufio"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/hash"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
)

// MemStorage stores metric info
type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

// NewMemStorage creates new MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		CounterMetrics: make(map[string]int64),
		GaugeMetrics:   make(map[string]float64),
	}
}

// FileStorage gets metric info from file and save metric info into file
type FileStorage struct {
	// StoreInterval is an interval in witch data is stored to file
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	storage       *MemStorage
}

// NewFileStorage creates new FileStorage
func NewFileStorage(storeFile string, storeInterval time.Duration, restore bool) FileStorage {
	return FileStorage{
		StoreInterval: storeInterval,
		StoreFile:     storeFile,
		Restore:       restore,
		storage:       NewMemStorage(),
	}
}

// storeMetricsToFile stores data from MemStorage to file
func (fs FileStorage) storeMetricsToFile() {
	file, err := os.OpenFile(fs.StoreFile, os.O_WRONLY|os.O_CREATE, 0777)
	writer := bufio.NewWriter(file)
	if err != nil {
		loggers.ErrorLogger.Printf("Failed to open file: %s", fs.StoreFile)
	}
	defer file.Close()
	for name, value := range fs.storage.CounterMetrics {
		gauge := types.Metrics{
			ID:    name,
			Delta: &value,
			MType: "counter",
		}
		jsonMetric, err := json.Marshal(gauge)
		if err != nil {
			loggers.ErrorLogger.Printf("error while marshalling json to file: %v", err)
		}
		_, err = writer.Write(jsonMetric)
		if err != nil {
			loggers.ErrorLogger.Printf("error while writing to file: %v", err)
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			loggers.ErrorLogger.Printf("error while writing counter to file: %v", err)
		}
	}
	for name, value := range fs.storage.GaugeMetrics {
		gauge := types.Metrics{
			ID:    name,
			Value: &value,
			MType: "gauge",
		}
		jsonMetric, err := json.Marshal(gauge)
		if err != nil {
			loggers.ErrorLogger.Printf("error while marshalling json to file: %v", err)
		}
		_, err = writer.Write(jsonMetric)
		if err != nil {
			loggers.ErrorLogger.Printf("error while writing gauge to file: %v", err)
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			loggers.ErrorLogger.Printf("error while writing gauge to file: %v", err)
		}
	}
	err = writer.Flush()
	if err != nil {
		loggers.ErrorLogger.Printf("writer.Flush() error: %v", err)
	}
	loggers.InfoLogger.Println("stored to file")
}

// RestoreMetrics restores metric from file to MemStorage
func (fs FileStorage) RestoreMetrics() error {
	file, err := os.OpenFile(fs.StoreFile, os.O_RDONLY|os.O_CREATE, 0777)
	scanner := bufio.NewScanner(file)
	if err != nil {
		loggers.ErrorLogger.Printf("Failed to open file: %s, %v", fs.StoreFile, err)
		return fmt.Errorf("error while opening file: %w", err)
	}
	defer file.Close()
	for scanner.Scan() {
		m := types.Metrics{}
		err = json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			loggers.ErrorLogger.Printf("json Unmarshal error: %v", err)
			return fmt.Errorf("error while unmarshalling json: %w", err)
		}
		fs.SaveMetric(m, "")
	}
	loggers.InfoLogger.Printf("Restored Metrics from '%s'", fs.StoreFile)
	return nil
}

// SaveMetric saves info about one metric into MemStorage
func (fs FileStorage) SaveMetric(m types.Metrics, key string) error {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("%wno value in update request", myerrors.ErrTypeNotImplemented)
		}
		if key != "" && m.Hash != "" {
			if !hmac.Equal([]byte(m.Hash), []byte(hash.Hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key))) {
				return fmt.Errorf("%wwrong hash in request", myerrors.ErrTypeBadRequest)
			}
		}
		fs.storage.GaugeMetrics[m.ID] = *m.Value
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("%wno value in update request", myerrors.ErrTypeNotImplemented)
		}
		if key != "" && m.Hash != "" {
			if !hmac.Equal([]byte(m.Hash), []byte(hash.Hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key))) {
				return fmt.Errorf("%wwrong hash in request", myerrors.ErrTypeBadRequest)
			}
		}
		fs.storage.CounterMetrics[m.ID] += *m.Delta
	default:
		return fmt.Errorf("%wno such type of metric", myerrors.ErrTypeNotImplemented)
	}
	return nil
}

// GetAllMetrics gets info about all metrics from MemStorage
func (fs FileStorage) GetAllMetrics() ([]types.Metrics, error) {
	var metrics []types.Metrics
	for name, value := range fs.storage.CounterMetrics {
		m := types.Metrics{ID: name, MType: "counter", Delta: &value}
		metrics = append(metrics, m)
	}
	for name, value := range fs.storage.GaugeMetrics {
		m := types.Metrics{ID: name, MType: "gauge", Value: &value}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// SetFileStorage file storage preferences
func (fs FileStorage) SetFileStorage() {
	if strings.LastIndex(fs.StoreFile, "/") != -1 {
		if err := os.MkdirAll(fs.StoreFile[:strings.LastIndex(fs.StoreFile, "/")], 0777); err != nil {
			loggers.ErrorLogger.Println("failed to create directory:", err)
		}
	}
	if fs.Restore {
		err := fs.RestoreMetrics()
		if err != nil {
			loggers.ErrorLogger.Println("error while restoring from file:", err)
		}
	}
	go repeating.Repeat(fs.storeMetricsToFile, fs.StoreInterval)
}

// GetMetric gets info about one metric from MemStorage
func (fs FileStorage) GetMetric(m types.Metrics, key string) (types.Metrics, error) {
	switch m.MType {
	case "counter":
		delta, ok := fs.storage.CounterMetrics[m.ID]
		if !ok {
			return m, myerrors.ErrTypeNotFound
		}
		m.Delta = &delta
	case "gauge":
		value, ok := fs.storage.GaugeMetrics[m.ID]
		if !ok {
			return m, myerrors.ErrTypeNotFound
		}
		m.Value = &value
	}
	return m, nil
}

// Check checks if file storage works OK
func (fs FileStorage) Check() error {
	return nil
}

// SaveManyMetrics saves several metrics into MemStorage
func (fs FileStorage) SaveManyMetrics(metrics []types.Metrics, key string) error {
	for _, m := range metrics {
		err := fs.SaveMetric(m, key)
		if err != nil {
			return err
		}
	}
	return nil
}
