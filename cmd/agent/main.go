package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
)

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultAddress        = "localhost:8080"
)

func main() {
	var reportInterval, pollInterval time.Duration
	if strPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		pollInterval = defaultPollInterval
	} else {
		var err error
		if pollInterval, err = time.ParseDuration(strPollInterval); err != nil {
			log.Println("couldn't parse poll duration")
		}
	}
	if strReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		reportInterval = defaultReportInterval
	} else {
		var err error
		if reportInterval, err = time.ParseDuration(strReportInterval); err != nil {
			log.Println("couldn't parse report duration")
		}
	}
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = defaultAddress
	}
	a := agent.NewAgent()
	a.Address = address
	go repeating.Repeat(a.CollectRuntimeMetrics, pollInterval)
	go repeating.Repeat(a.SendAllMetrics, reportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
