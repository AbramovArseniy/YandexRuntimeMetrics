package server

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"
	"strconv"
	"strings"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/go-chi/chi/v5"
)

const contentTypeJSON = "application/json"

var ErrTypeNotImplemented = errors.New("")

func CompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func DecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			rw.Header().Set("Accept-Encoding", "gzip")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
			next.ServeHTTP(rw, r)
		} else {
			next.ServeHTTP(rw, r)
		}
	})

}

func GetGaugeStatusOK(rw http.ResponseWriter, metricVal float64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
	_, err := rw.Write([]byte(strVal))
	if err != nil {
		loggers.ErrorLogger.Println("response writer error:", err)
		return
	}
}

func GetCounterStatusOK(rw http.ResponseWriter, metricVal int64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	_, err := rw.Write([]byte(fmt.Sprintf("%d", metricVal)))
	if err != nil {
		loggers.ErrorLogger.Println("response writer error:", err)
		return
	}

}

func (s *Server) GetAllMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	loggers.InfoLogger.Println("Get all request")
	rw.Header().Set("Content-Type", "text/html")
	for metricName, metricVal := range s.storage.GaugeMetrics {
		strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
		_, err := rw.Write([]byte(fmt.Sprintf("%s: %s\n", metricName, strVal)))
		if err != nil {
			loggers.ErrorLogger.Println("response writer error:", err)
			return
		}
	}
	for metricName, metricVal := range s.storage.CounterMetrics {
		_, err := rw.Write([]byte(fmt.Sprintf("%s: %d", metricName, metricVal)))
		if err != nil {
			loggers.ErrorLogger.Println("response writer error:", err)
			return
		}
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value")
	switch metricType {
	case "gauge":
		newVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "Wrong Gauge Value", http.StatusBadRequest)
		}
		s.storage.GaugeMetrics[metricName] = newVal
	case "counter":
		newVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "Wrong Counter Value", http.StatusBadRequest)
		}
		s.storage.CounterMetrics[metricName] += newVal
	default:
		loggers.ErrorLogger.Printf("wrong Metric Type: %s", metricType)
		http.Error(rw, "Wrong Metric Type", http.StatusNotImplemented)
	}
	if s.Debug {
		loggers.DebugLogger.Printf("POST %s %s", metricType, metricName)
	}
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName := chi.URLParam(r, "type"), chi.URLParam(r, "name")
	if s.Debug {
		loggers.DebugLogger.Printf("GET %s %s", metricType, metricName)
	}
	switch metricType {
	case "gauge":
		if metricVal, isIn := s.storage.GaugeMetrics[metricName]; isIn {
			GetGaugeStatusOK(rw, metricVal)
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}

	case "counter":
		if metricVal, isIn := s.storage.CounterMetrics[metricName]; isIn {
			GetCounterStatusOK(rw, metricVal)
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}
	default:
		http.Error(rw, "There is no metric you requested", http.StatusNotFound)
	}
}

func (s *Server) PostMetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
			return
		}
		loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
		return
	}
	var m Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		loggers.ErrorLogger.Printf("Decode error: %v", err)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			return
		}
		return
	}
	if s.Debug {
		loggers.DebugLogger.Println("POST JSON " + m.ID + " " + m.MType)
	}
	err := s.storeMetrics(m)
	if err != nil {
		rw.Header().Set("Content-Type", contentTypeJSON)
		loggers.ErrorLogger.Println(err.Error())
		if errors.Is(err, ErrTypeNotImplemented) {
			http.Error(rw, err.Error(), http.StatusNotImplemented)
		} else {
			loggers.ErrorLogger.Println("Store Metrics error:", err.Error())
		}
		return

	}
	rw.Header().Add("Content-Type", contentTypeJSON)
	jsonMetric, err := json.Marshal(m)
	if err != nil {
		loggers.ErrorLogger.Printf("json Marshal error: %s", err)
		return
	}
	_, err = rw.Write(jsonMetric)
	if err != nil {
		http.Error(rw, "can't write body", http.StatusInternalServerError)
		loggers.ErrorLogger.Println(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricPostJSONHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
			return
		}
		return
	}
	body, err := io.ReadAll(r.Body)
	rw.Header().Set("Content-Type", "application/json")

	if err != nil {
		http.Error(rw, "reading body error", http.StatusInternalServerError)
		return
	}
	var m Metrics
	if err = json.Unmarshal(body, &m); err != nil {
		http.Error(rw, "Could not unmarshal JSON:", http.StatusInternalServerError)
		return
	}

	switch m.MType {
	case "counter":
		val, isIn := s.storage.CounterMetrics[m.ID]
		if !isIn {
			if s.Debug {
				loggers.DebugLogger.Println("There is no metric you requested")
			}
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
			return
		}
		m.Delta = &val
	case "gauge":
		val, isIn := s.storage.GaugeMetrics[m.ID]
		if !isIn {
			if s.Debug {
				loggers.DebugLogger.Println("There is no metric you requested")
			}
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
			return
		}
		m.Value = &val
	default:
		if s.Debug {
			loggers.DebugLogger.Println("There is no metric you requested")
		}
		http.Error(rw, "There is no metric You requested", http.StatusNotFound)
		return
	}
	jsonMetric, err := json.Marshal(m)
	if err != nil {
		loggers.ErrorLogger.Printf("json Marshal error: %s", err)
		return
	}
	_, err = rw.Write(jsonMetric)
	if err != nil {
		http.Error(rw, "can't write body", http.StatusInternalServerError)
		loggers.ErrorLogger.Println(err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) storeMetrics(m Metrics) error {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		s.storage.GaugeMetrics[m.ID] = *m.Value
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		s.storage.CounterMetrics[m.ID] += *m.Delta
	default:
		return fmt.Errorf("%wno such type of metric", ErrTypeNotImplemented)
	}
	return nil
}
