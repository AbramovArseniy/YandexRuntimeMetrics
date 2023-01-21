package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
)

func StartServer() {
	s := server.NewServer()
	srv := &http.Server{
		Handler: s.Router(),
	}
	addr, exists := os.LookupEnv("ADDRESS")
	if !exists {
		srv.Addr = defaultAddress
	} else {
		srv.Addr = addr
	}
	var storeInterval time.Duration
	s.FileHandler = server.FileHandler{}
	err := env.Parse(&s.FileHandler)
	log.Println(s.FileHandler)
	if err != nil {
		log.Println("error parsing env")
		return
	}
	if s.FileHandler.Restore {
		s.RestoreMetricsFromFile()
	}
	log.Println("Server started")
	go repeating.Repeat(s.StoreMetricsToFile, storeInterval*time.Second)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-cancelSignal
}

func main() {
	StartServer()
}
