// Package main starts agent
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shirou/gopsutil/cpu"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/types"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
)

// build info
var buildVersion, buildDate, buildCommit string = "N/A", "N/A", "N/A"

// main starts agent
func main() {
	a, err := agent.NewAgent(config.SetAgentParams())
	if err != nil {
		loggers.ErrorLogger.Fatal("wrong protocol")
	}
	cpuStat, err := cpu.Times(true)
	if err != nil {
		log.Println(err)
		return
	}
	numCPU := len(cpuStat)
	a.Collector.UtilData.CPUtime = make([]float64, numCPU)
	a.Collector.UtilData.CPUutilizations = make([]types.Metrics, numCPU)
	loggers.InfoLogger.Printf(`Build version: %s
	Build date: %s
	Build commit: %s`,
		buildVersion, buildDate, buildCommit)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go repeating.Repeat(sigs, a.Collector.CollectRuntimeMetrics, a.PollInterval)
	go repeating.Repeat(sigs, a.Collector.CollectUtilizationMetrics, a.PollInterval)
	go repeating.Repeat(sigs, a.SendAllMetricsAsButch, a.ReportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
