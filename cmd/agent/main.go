package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent"
)

const (
	pollRuntimeMetricsInterval = 2 * time.Second
	reportInterval             = 10 * time.Second
)

func main() {
	var a agent.Agent = *agent.NewAgent()
	go agent.Repeat(a.CollectRuntimeMetrics, pollRuntimeMetricsInterval)
	go agent.Repeat(a.SendAllMetrics, reportInterval)
	log.Println("Agent started")
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}
