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

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultAddress        = "localhost:8080"
)

func main() {
	a := agent.NewAgent()
	var (
		flagPollInterval   time.Duration
		flagReportInterval time.Duration
		flagAddress        string
	)
	flag.DurationVar(&flagPollInterval, "p", defaultPollInterval, "poll_metrics_interval")
	flag.DurationVar(&flagReportInterval, "r", defaultReportInterval, "report_metrics_interval")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.Parse()
	if strPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		a.PollInterval = flagPollInterval
	} else {
		var err error
		if a.PollInterval, err = time.ParseDuration(strPollInterval); err != nil || a.PollInterval <= 0 {
			log.Println("couldn't parse poll duration from environment")
			a.PollInterval = flagPollInterval
		}
	}
	if strReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		a.ReportInterval = flagReportInterval
	} else {
		var err error
		if a.ReportInterval, err = time.ParseDuration(strReportInterval); err != nil || a.ReportInterval <= 0 {
			log.Println("couldn't parse report duration from")
			a.ReportInterval = flagReportInterval
		}
	}
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = flagAddress
	}
	a.Address = address
	go repeating.Repeat(a.CollectRuntimeMetrics, a.PollInterval)
	go repeating.Repeat(a.SendAllMetrics, a.ReportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
