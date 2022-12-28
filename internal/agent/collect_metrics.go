package agent

import (
	"math/rand"
	"runtime"
	"time"
)

func CollectRuntimeMetrics() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	PollCount += 1
	Metrics = []Gauge{
		{metricName: "Alloc", metricValue: float64(stats.Alloc)},
		{metricName: "BuckHashSys", metricValue: float64(stats.BuckHashSys)},
		{metricName: "Frees", metricValue: float64(stats.Frees)},
		{metricName: "GCCPUFraction", metricValue: stats.GCCPUFraction},
		{metricName: "GCSys", metricValue: float64(stats.GCSys)},
		{metricName: "HeapAlloc", metricValue: float64(stats.HeapAlloc)},
		{metricName: "HeapIdle", metricValue: float64(stats.HeapIdle)},
		{metricName: "HeapInuse", metricValue: float64(stats.HeapInuse)},
		{metricName: "HeapObjects", metricValue: float64(stats.HeapObjects)},
		{metricName: "HeapReleased", metricValue: float64(stats.HeapReleased)},
		{metricName: "HeapSys", metricValue: float64(stats.HeapSys)},
		{metricName: "LastGC", metricValue: float64(stats.LastGC)},
		{metricName: "Lookups", metricValue: float64(stats.Lookups)},
		{metricName: "MCacheInuse", metricValue: float64(stats.MCacheInuse)},
		{metricName: "MCacheSys", metricValue: float64(stats.MCacheSys)},
		{metricName: "MSpanInuse", metricValue: float64(stats.MSpanInuse)},
		{metricName: "MSpanSys", metricValue: float64(stats.MSpanSys)},
		{metricName: "Mallocs", metricValue: float64(stats.Mallocs)},
		{metricName: "NextGC", metricValue: float64(stats.NextGC)},
		{metricName: "NumForcedGC", metricValue: float64(stats.NumForcedGC)},
		{metricName: "NumGC", metricValue: float64(stats.NumGC)},
		{metricName: "OtherSys", metricValue: float64(stats.OtherSys)},
		{metricName: "PauseTotalNs", metricValue: float64(stats.PauseTotalNs)},
		{metricName: "StackInuse", metricValue: float64(stats.StackInuse)},
		{metricName: "StackSys", metricValue: float64(stats.StackSys)},
		{metricName: "Sys", metricValue: float64(stats.Sys)},
	}
}

func CollectRandomValueMetric(metrics []Gauge) []Gauge {
	rand.Seed(time.Now().Unix())
	randomValueMetric := Gauge{metricName: "RandomValue", metricValue: rand.Float64() * 1000}
	metrics = append(metrics, randomValueMetric)
	return metrics
}
