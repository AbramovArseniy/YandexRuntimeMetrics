package main

import (
	"flag"
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
	handler := server.DecompressHandler(s.Router())
	handler = server.CompressHandler(handler)
	srv := &http.Server{
		Handler: handler,
	}
	var (
		flagRestore       bool
		flagStoreFile     string
		flagAddress       string
		flagStoreInterval time.Duration
	)
	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore_true/false")
	flag.StringVar(&flagStoreFile, "f", defaultStoreFile, "store_file")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.DurationVar(&flagStoreInterval, "i", defaultStoreInterval, "store_interval_in_seconds")
	flag.Parse()
	addr, exists := os.LookupEnv("ADDRESS")
	if !exists {
		srv.Addr = flagAddress
	} else {
		srv.Addr = addr
	}
	if s.FileHandler.StoreFile, exists = os.LookupEnv("STORE_FILE"); !exists {
		s.FileHandler.StoreFile = flagStoreFile
	}
	if strStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		s.FileHandler.StoreInterval = flagStoreInterval
	} else {
		var err error
		if s.FileHandler.StoreInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			log.Println("couldn't parse store interval")
			s.FileHandler.StoreInterval = flagStoreInterval
		}
	}
	if strRestore, exists := os.LookupEnv("RESTORE"); !exists {
		s.FileHandler.Restore = flagRestore
	} else {
		var err error
		if s.FileHandler.Restore, err = strconv.ParseBool(strRestore); err != nil {
			log.Println("couldn't parse restore bool")
			s.FileHandler.Restore = flagRestore
		}
	}
	if strings.LastIndex(s.FileHandler.StoreFile, "/") != -1 {
		if err := os.MkdirAll(s.FileHandler.StoreFile[:strings.LastIndex(s.FileHandler.StoreFile, "/")], 0777); err != nil {
			log.Println("failed to create directory:", err)
		}
	}
	if s.FileHandler.Restore {
		s.RestoreMetricsFromFile()
	}
	log.Println("Server started")
	go repeating.Repeat(s.StoreMetricsToFile, s.FileHandler.StoreInterval)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func main() {
	StartServer()
}
