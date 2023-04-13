// Package main starts agent
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
)

// default agent preferences
const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultAddress        = "localhost:8080"
	defaultRateLimit      = 100
)

// setAgentParams set agent config
func setAgentParams() (string, time.Duration, time.Duration, string, int) {
	var (
		flagPollInterval   time.Duration
		flagReportInterval time.Duration
		flagRateLimit      int
		flagAddress        string
		flagKey            string
	)
	flag.DurationVar(&flagPollInterval, "p", defaultPollInterval, "poll_metrics_interval")
	flag.DurationVar(&flagReportInterval, "r", defaultReportInterval, "report_metrics_interval")
	flag.IntVar(&flagRateLimit, "l", defaultRateLimit, "rate_limit")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.StringVar(&flagKey, "k", "", "hash_key")
	flag.Parse()
	var pollInterval, reportInterval time.Duration
	if strPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		pollInterval = flagPollInterval
	} else {
		var err error
		if pollInterval, err = time.ParseDuration(strPollInterval); err != nil || pollInterval <= 0 {
			log.Println("couldn't parse poll duration from environment")
			pollInterval = flagPollInterval
		}
	}
	if strReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		reportInterval = flagReportInterval
	} else {
		var err error
		if reportInterval, err = time.ParseDuration(strReportInterval); err != nil || reportInterval <= 0 {
			log.Println("couldn't parse report duration from")
			reportInterval = flagReportInterval
		}
	}
	var rateLimit int
	if strRateLimit, exists := os.LookupEnv("RATE_LIMIT"); !exists {
		rateLimit = flagRateLimit
	} else {
		var err error
		if rateLimit, err = strconv.Atoi(strRateLimit); err != nil || rateLimit <= 0 {
			log.Println("couldn't parse report duration from", strRateLimit)
			rateLimit = flagRateLimit
		}
	}
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = flagAddress
	}
	key, exists := os.LookupEnv("KEY")
	if !exists {
		key = flagKey
	}
	return address, pollInterval, reportInterval, key, rateLimit
}

// main starts agent
func main() {
	a := agent.NewAgent(setAgentParams())
	cpuStat, err := cpu.Times(true)
	if err != nil {
		log.Println(err)
		return
	}
	numCPU := len(cpuStat)
	a.UtilData.CPUtime = make([]float64, numCPU)
	a.UtilData.CPUutilizations = make([]agent.Metrics, numCPU)
	go repeating.Repeat(a.CollectRuntimeMetrics, a.PollInterval)
	go repeating.Repeat(a.CollectUtilizationMetrics, a.PollInterval)
	go repeating.Repeat(a.SendAllMetricsAsButch, a.ReportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
