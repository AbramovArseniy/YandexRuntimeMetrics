package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
)

func setAgentParams() (string, time.Duration, time.Duration, string) {
	var (
		flagPollInterval   time.Duration
		flagReportInterval time.Duration
		flagAddress        string
		flagKey            string
	)
	flag.DurationVar(&flagPollInterval, "p", defaultPollInterval, "poll_metrics_interval")
	flag.DurationVar(&flagReportInterval, "r", defaultReportInterval, "report_metrics_interval")
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
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = flagAddress
	}
	key, exists := os.LookupEnv("KEY")
	if !exists {
		key = flagKey
	}
	return address, pollInterval, reportInterval, key
}

func main() {
	a := agent.NewAgent(setAgentParams())
	go repeating.Repeat(a.CollectRuntimeMetrics, a.PollInterval)
	go repeating.Repeat(a.SendAllMetricsAsButch, a.ReportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
