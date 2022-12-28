package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	metricType, metricName, metricValue := strings.Split(fmt.Sprintf("%v", r.URL), "/")[2], strings.Split(fmt.Sprintf("%v", r.URL), "/")[3], strings.Split(fmt.Sprintf("%v", r.URL), "/")[4]
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
		newVal, err := strconv.ParseInt(metricValue, 32, 64)
		if err != nil {
			log.Println()
			http.Error(rw, "Wrong Counter Value", http.StatusBadRequest)
		}
		log.Printf("%s: %d", metricName, newVal)
		Storage.CounterMetrics[metricName] += newVal
	}
	log.Println(Storage)

}
