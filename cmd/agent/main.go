package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbramovArseniy/RuntimeMetrics/internal/agent"
)

const (
	pollRuntimeMetricsInterval = 2 * time.Second
	reportInterval             = 10 * time.Second
)

func main() {
	go agent.Repeat(agent.CollectRuntimeMetrics, pollRuntimeMetricsInterval)
	go agent.Repeat(agent.SendAllMetrics, reportInterval)

	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal

	os.Exit(1)
}
