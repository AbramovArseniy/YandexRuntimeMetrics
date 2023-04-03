package agent

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// CollectRuntimeMetrics collects runtime metrics
func (a *Agent) CollectRuntimeMetrics() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	var (
		Alloc         = float64(stats.Alloc)
		BuckHashSys   = float64(stats.BuckHashSys)
		Frees         = float64(stats.Frees)
		GCCPUFraction = float64(stats.GCCPUFraction)
		GCSys         = float64(stats.GCSys)
		HeapAlloc     = float64(stats.HeapAlloc)
		HeapIdle      = float64(stats.HeapIdle)
		HeapInuse     = float64(stats.HeapInuse)
		HeapObjects   = float64(stats.HeapObjects)
		HeapReleased  = float64(stats.HeapReleased)
		HeapSys       = float64(stats.HeapSys)
		LastGC        = float64(stats.LastGC)
		Lookups       = float64(stats.Lookups)
		MCacheInuse   = float64(stats.MCacheInuse)
		MCacheSys     = float64(stats.MCacheSys)
		MSpanInuse    = float64(stats.MSpanInuse)
		MSpanSys      = float64(stats.MSpanSys)
		Mallocs       = float64(stats.Mallocs)
		NextGC        = float64(stats.NextGC)
		NumForcedGC   = float64(stats.NumForcedGC)
		NumGC         = float64(stats.NumGC)
		OtherSys      = float64(stats.OtherSys)
		PauseTotalNs  = float64(stats.PauseTotalNs)
		StackInuse    = float64(stats.StackInuse)
		StackSys      = float64(stats.StackSys)
		Sys           = float64(stats.Sys)
		TotalAlloc    = float64(stats.TotalAlloc)
	)
	a.collector.RuntimeMetrics = []Metrics{
		{ID: "Alloc", MType: "gauge", Value: &Alloc},
		{ID: "BuckHashSys", MType: "gauge", Value: &BuckHashSys},
		{ID: "Frees", MType: "gauge", Value: &Frees},
		{ID: "GCCPUFraction", MType: "gauge", Value: &GCCPUFraction},
		{ID: "GCSys", MType: "gauge", Value: &GCSys},
		{ID: "HeapAlloc", MType: "gauge", Value: &HeapAlloc},
		{ID: "HeapIdle", MType: "gauge", Value: &HeapIdle},
		{ID: "HeapInuse", MType: "gauge", Value: &HeapInuse},
		{ID: "HeapObjects", MType: "gauge", Value: &HeapObjects},
		{ID: "HeapReleased", MType: "gauge", Value: &HeapReleased},
		{ID: "HeapSys", MType: "gauge", Value: &HeapSys},
		{ID: "LastGC", MType: "gauge", Value: &LastGC},
		{ID: "Lookups", MType: "gauge", Value: &Lookups},
		{ID: "MCacheInuse", MType: "gauge", Value: &MCacheInuse},
		{ID: "MCacheSys", MType: "gauge", Value: &MCacheSys},
		{ID: "MSpanInuse", MType: "gauge", Value: &MSpanInuse},
		{ID: "MSpanSys", MType: "gauge", Value: &MSpanSys},
		{ID: "Mallocs", MType: "gauge", Value: &Mallocs},
		{ID: "NextGC", MType: "gauge", Value: &NextGC},
		{ID: "NumForcedGC", MType: "gauge", Value: &NumForcedGC},
		{ID: "NumGC", MType: "gauge", Value: &NumGC},
		{ID: "OtherSys", MType: "gauge", Value: &OtherSys},
		{ID: "PauseTotalNs", MType: "gauge", Value: &PauseTotalNs},
		{ID: "StackInuse", MType: "gauge", Value: &StackInuse},
		{ID: "StackSys", MType: "gauge", Value: &StackSys},
		{ID: "Sys", MType: "gauge", Value: &Sys},
		{ID: "TotalAlloc", MType: "gauge", Value: &TotalAlloc},
	}
	*(a.collector.PollCount.Delta)++
	loggers.InfoLogger.Println("Collected GaugeMetrics")
}

// CollectRandomValueMetric collects metric with random value
func (s *metricCollector) CollectRandomValueMetric() Metrics {
	rand.Seed(time.Now().Unix())
	value := rand.Float64() * 1000
	randomValueMetric := Metrics{ID: "RandomValue", MType: "gauge", Value: &value}
	loggers.InfoLogger.Println("Collected RandomValueMectric")
	return randomValueMetric
}

// CollectUtilizationMetrics collects cpu and mem metrics
func (a *Agent) CollectUtilizationMetrics() {
	m, err := mem.VirtualMemory()
	if err != nil {
		loggers.ErrorLogger.Println("error access to virtual memory: ", err)
	}

	a.UtilData.mu.Lock()
	timeNow := time.Now()
	timeDiff := timeNow.Sub(a.UtilData.CPUutilLastTime)
	Total := float64(m.Total)
	Free := float64(m.Free)
	a.UtilData.CPUutilLastTime = timeNow
	a.UtilData.TotalMemory = Metrics{
		ID:    "TotalMemory",
		MType: "gauge",
		Value: &Total,
	}
	a.UtilData.FreeMemory = Metrics{
		ID:    "FreeMemory",
		MType: "gauge",
		Value: &Free,
	}

	cpus, err := cpu.Times(true)
	if err != nil {
		log.Println(err)
	}
	for i := range cpus {
		newCPUTime := cpus[i].User + cpus[i].System
		cpuUtilization := (newCPUTime - a.UtilData.CPUtime[i]) * 1000 / float64(timeDiff.Milliseconds())
		a.UtilData.CPUutilizations[i] = Metrics{
			ID:    "CPUutilization" + strconv.Itoa(i+1),
			MType: "gauge",
			Value: &cpuUtilization,
		}
		a.UtilData.CPUtime[i] = newCPUTime
	}
	a.UtilData.mu.Unlock()
}
