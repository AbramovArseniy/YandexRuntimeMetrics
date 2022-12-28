package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func GetAllMetricsHandler(rw http.ResponseWriter, _ *http.Request) {
	log.Println("Get all request")
	for metricName, metricVal := range Storage.GaugeMetrics {
		strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
		rw.Write([]byte(fmt.Sprintf("%s: %s\n", metricName, strVal)))
	}
	for metricName, metricVal := range Storage.CounterMetrics {
		rw.Write([]byte(fmt.Sprintf("%s: %d", metricName, metricVal)))
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
}

func PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value")
	log.Println(metricType, metricName, metricValue)
	switch metricType {
	case "gauge":
		newVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Println()
			http.Error(rw, "Wrong Gauge Value", http.StatusBadRequest)
		}
		log.Printf("%s: %f", metricName, newVal)
		Storage.GaugeMetrics[metricName] = newVal
	case "counter":
		newVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			log.Println()
			http.Error(rw, "Wrong Counter Value", http.StatusBadRequest)
		}
		log.Printf("%s: %d", metricName, newVal)
		Storage.CounterMetrics[metricName] += newVal
	default:
		log.Printf("wrong Metric Type: %s", metricType)
		http.Error(rw, "Wrong Metric Type", http.StatusNotImplemented)
	}
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

}

func GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName := chi.URLParam(r, "type"), chi.URLParam(r, "name")
	log.Println(metricType, metricName)
	switch metricType {
	case "gauge":
		log.Printf("GetRequest: value of %s", metricName)
		if metricVal, isIn := Storage.GaugeMetrics[metricName]; isIn {
			rw.WriteHeader(http.StatusOK)
			rw.Header().Add("Content-Type", "text/plain")
			strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
			_, err := rw.Write([]byte(strVal))
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}

	case "counter":
		log.Printf("GetRequest: value of %s", metricName)
		if metricVal, isIn := Storage.CounterMetrics[metricName]; isIn {
			rw.WriteHeader(http.StatusOK)
			rw.Header().Add("Content-Type", "text/plain")
			_, err := rw.Write([]byte(fmt.Sprintf("%d", metricVal)))
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		}
	default:
		http.Error(rw, "There is no metric you requested", http.StatusNotFound)
	}
	log.Println(Storage)
}
