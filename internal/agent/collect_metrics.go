package agent

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

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
		{ID: "Alloc", Value: &Alloc},
		{ID: "BuckHashSys", Value: &BuckHashSys},
		{ID: "Frees", Value: &Frees},
		{ID: "GCCPUFraction", Value: &GCCPUFraction},
		{ID: "GCSys", Value: &GCSys},
		{ID: "HeapAlloc", Value: &HeapAlloc},
		{ID: "HeapIdle", Value: &HeapIdle},
		{ID: "HeapInuse", Value: &HeapInuse},
		{ID: "HeapObjects", Value: &HeapObjects},
		{ID: "HeapReleased", Value: &HeapReleased},
		{ID: "HeapSys", Value: &HeapSys},
		{ID: "LastGC", Value: &LastGC},
		{ID: "Lookups", Value: &Lookups},
		{ID: "MCacheInuse", Value: &MCacheInuse},
		{ID: "MCacheSys", Value: &MCacheSys},
		{ID: "MSpanInuse", Value: &MSpanInuse},
		{ID: "MSpanSys", Value: &MSpanSys},
		{ID: "Mallocs", Value: &Mallocs},
		{ID: "NextGC", Value: &NextGC},
		{ID: "NumForcedGC", Value: &NumForcedGC},
		{ID: "NumGC", Value: &NumGC},
		{ID: "OtherSys", Value: &OtherSys},
		{ID: "PauseTotalNs", Value: &PauseTotalNs},
		{ID: "StackInuse", Value: &StackInuse},
		{ID: "StackSys", Value: &StackSys},
		{ID: "Sys", Value: &Sys},
		{ID: "TotalAlloc", Value: &TotalAlloc},
	}
	*(a.collector.PollCount.Delta)++
	loggers.InfoLogger.Println("Collected GaugeMetrics")
}

func (s *metricCollector) CollectRandomValueMetric() Metrics {
	rand.Seed(time.Now().Unix())
	value := rand.Float64() * 1000
	randomValueMetric := Metrics{ID: "RandomValue", Value: &value}
	loggers.InfoLogger.Println("Collected RandomValueMectric")
	return randomValueMetric
}

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
