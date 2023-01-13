package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const contentTypeJson = "application/json"

type Handler struct {
	storage MemStorage
}

func GetGaugeStatusOK(rw http.ResponseWriter, metricVal float64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
	_, err := rw.Write([]byte(strVal))
	if err != nil {
		log.Println(err)
		return
	}
}

func GetCounterStatusOK(rw http.ResponseWriter, metricVal int64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	_, err := rw.Write([]byte(fmt.Sprintf("%d", metricVal)))
	if err != nil {
		log.Println(err)
		return
	}
}

func (h *Handler) GetAllMetricsHandler(rw http.ResponseWriter, _ *http.Request) {
	log.Println("Get all request")
	for metricName, metricVal := range h.storage.GaugeMetrics {
		strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
		rw.Write([]byte(fmt.Sprintf("%s: %s\n", metricName, strVal)))
	}
	for metricName, metricVal := range h.storage.CounterMetrics {
		rw.Write([]byte(fmt.Sprintf("%s: %d", metricName, metricVal)))
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
}

func (h *Handler) PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value")
	switch metricType {
	case "gauge":
		newVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "Wrong Gauge Value", http.StatusBadRequest)
		}
		h.storage.GaugeMetrics[metricName] = newVal
	case "counter":
		newVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "Wrong Counter Value", http.StatusBadRequest)
		}
		h.storage.CounterMetrics[metricName] += newVal
	default:
		log.Printf("wrong Metric Type: %s", metricType)
		http.Error(rw, "Wrong Metric Type", http.StatusNotImplemented)
	}
	log.Printf("POST %s %s", metricType, metricName)
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName := chi.URLParam(r, "type"), chi.URLParam(r, "name")
	log.Printf("GET %s %s", metricType, metricName)
	switch metricType {
	case "gauge":
		if metricVal, isIn := h.storage.GaugeMetrics[metricName]; isIn {
			GetGaugeStatusOK(rw, metricVal)
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}

	case "counter":
		if metricVal, isIn := h.storage.CounterMetrics[metricName]; isIn {
			GetCounterStatusOK(rw, metricVal)
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}
	default:
		http.Error(rw, "There is no metric you requested", http.StatusNotFound)
	}
}

func (h *Handler) PostMetricJsonHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJson {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			log.Println("Wrong content type")
			return
		}
		log.Println("Wrong content type")
		return
	}
	log.Println(r.Body)
	var m Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("Decode problem")
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			return
		}
		return
	}
	log.Println("POST JSON" + m.ID + m.MType)
	err := h.storeMetrics(m)
	if err != nil {
		rw.Header().Set("Content-Type", contentTypeJson)
		log.Println(err)
		http.Error(rw, "No such type of metric", http.StatusNotImplemented)
		return
	}
	rw.Header().Add("Content-Type", contentTypeJson)
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricPostJsonHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJson {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			log.Println("Wrong content type")
			return
		}
		return
	}
	log.Println(r.Body)
	var m Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			log.Println("Decode problem")
			return
		}
		return
	}
	switch m.MType {
	case "counter":
		*m.Delta = h.storage.CounterMetrics[m.ID]
	case "gauge":
		*m.Value = h.storage.GaugeMetrics[m.ID]
	default:
		http.Error(rw, "There is no metric you requested", http.StatusNotFound)
	}
	var jsonMetric, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Printf("json Marshal error: %s", err)
		return
	}
	_, err = rw.Write(jsonMetric)
	if err != nil {
		log.Println(err)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) storeMetrics(m Metrics) error {
	switch m.MType {
	case "gauge":
		log.Printf("saving metric %s %s %f\n", m.ID, m.MType, *m.Value)
		h.storage.GaugeMetrics[m.ID] = *m.Value
	case "counter":
		log.Printf("saving metric %s %s %d\n", m.ID, m.MType, *m.Delta)
		h.storage.CounterMetrics[m.ID] += *m.Delta
	default:
		return errors.New("no such type of metric")
	}
	return nil
}
