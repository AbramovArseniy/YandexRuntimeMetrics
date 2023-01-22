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
	defaultStoreInterval = 300
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
)

func initFlags(s *server.Server) {
	flag.BoolVar(&s.FileHandler.Restore, "r", true, "Restore")
	flag.StringVar(&s.FileHandler.StoreFile, "f", "/tmp/devops-metrics-db.json", "storeFile")
	flag.StringVar(&s.Addr, "a", "localhost:8080", "address")
	flag.IntVar(&s.FileHandler.StoreInterval, "i", 300, "time_in_seconds")
	flag.Parse()
}

func StartServer() {
	s := server.NewServer()
	initFlags(s)
	addr, exists := os.LookupEnv("ADDRESS")
	if !exists {
		s.Addr = defaultAddress
	} else {
		s.Addr = addr
	}
	srv := &http.Server{
		Handler: s.Router(),
		Addr:    s.Addr,
	}
	if s.FileHandler.StoreFile, exists = os.LookupEnv("STORE_FILE"); !exists {
		flag.Parse()
	}
	if strStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		flag.Parse()
	} else {
		var err error
		if s.FileHandler.StoreInterval, err = strconv.Atoi(strStoreInterval); err != nil {
			log.Println("couldn't parse store interval")
			flag.Parse()
		}
	}
	if strRestore, exists := os.LookupEnv("RESTORE"); !exists {
		flag.Parse()
	} else {
		var err error
		if s.FileHandler.Restore, err = strconv.ParseBool(strRestore); err != nil {
			log.Println("couldn't parse restore bool")
			flag.Parse()
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
	go repeating.Repeat(s.StoreMetricsToFile, time.Duration(s.FileHandler.StoreInterval)*time.Second)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func main() {
	StartServer()
}
