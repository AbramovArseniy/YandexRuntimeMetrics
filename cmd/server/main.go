package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
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
	if s.StoreFile, exists = os.LookupEnv("ADDRESS"); !exists {
		s.StoreFile = defaultStoreFile
	}
	if strStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		storeInterval = defaultStoreInterval
	} else {
		var err error
		if storeInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			log.Println("couldn't parse store interval")
			storeInterval = defaultStoreInterval
		}
	}
	var Restore bool
	if strRestore, exists := os.LookupEnv("RESTORE"); !exists {
		Restore = defaultRestore
	} else {
		var err error
		if Restore, err = strconv.ParseBool(strRestore); err != nil {
			log.Println("couldn't parse restore bool")
			Restore = defaultRestore
		}
	}
	if err := os.MkdirAll(s.StoreFile[:strings.LastIndex(s.StoreFile, "/")], 0777); err != nil {
		log.Println("failed to create directory:", err)
	}
	if Restore {
		s.RestoreMetricsFromFile()
	}
	log.Println("Server started")
	go repeating.Repeat(s.StoreMetricsToFile, storeInterval)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func main() {
	StartServer()
}
