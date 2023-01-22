package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
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

func initFlags(a *agent.Agent) {
	flag.StringVar(&a.Address, "a", "localhost:8080", "address")
	flag.IntVar(&a.PollInterval, "p", 2, "poll_interval")
	flag.IntVar(&a.ReportInterval, "r", 10, "report_interval")
}

func main() {
	a := agent.NewAgent()
	initFlags(a)
	if strPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		flag.Parse()
	} else {
		var err error
		if a.PollInterval, err = strconv.Atoi(strPollInterval); err != nil {
			log.Println("couldn't parse poll duration from environment")
			flag.Parse()
		}
	}
	if strReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		flag.Parse()
	} else {
		var err error
		if a.ReportInterval, err = strconv.Atoi(strReportInterval); err != nil {
			log.Println("couldn't parse report duration from")
			flag.Parse()
		}
	}
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = defaultAddress
	}
	a.Address = address
	go repeating.Repeat(a.CollectRuntimeMetrics, time.Duration(a.PollInterval)*time.Second)
	go repeating.Repeat(a.SendAllMetrics, time.Duration(a.ReportInterval)*time.Second)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
