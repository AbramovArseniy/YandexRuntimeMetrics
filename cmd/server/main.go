package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/repeating"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
)

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
)

func setServerParams() (string, time.Duration, string, bool, bool) {
	var (
		flagRestore, restore             bool
		flagStoreFile, storeFile         string
		flagAddress                      string
		flagStoreInterval, storeInterval time.Duration
		flagDebug                        bool
	)

	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore_true/false")
	flag.StringVar(&flagStoreFile, "f", defaultStoreFile, "store_file")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.DurationVar(&flagStoreInterval, "i", defaultStoreInterval, "store_interval_in_seconds")
	flag.BoolVar(&flagDebug, "d", false, "debug_true/false")
	flag.Parse()
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = flagAddress
	}
	if storeFile, exists = os.LookupEnv("STORE_FILE"); !exists {
		storeFile = flagStoreFile
	}
	if strStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		storeInterval = flagStoreInterval
	} else {
		var err error
		if storeInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			loggers.ErrorLogger.Println("couldn't parse store interval")
			storeInterval = flagStoreInterval
		}
	}
	if strRestore, exists := os.LookupEnv("RESTORE"); !exists {
		restore = flagRestore
	} else {
		var err error
		if restore, err = strconv.ParseBool(strRestore); err != nil {
			loggers.ErrorLogger.Println("couldn't parse restore bool")
			restore = flagRestore
		}
	}
	return address, storeInterval, storeFile, restore, flagDebug
}

func StartServer() {
	s := server.NewServer(setServerParams())
	handler := server.DecompressHandler(s.Router())
	handler = server.CompressHandler(handler)
	srv := &http.Server{
		Addr:    s.Addr,
		Handler: handler,
	}
	if strings.LastIndex(s.FileHandler.StoreFile, "/") != -1 {
		if err := os.MkdirAll(s.FileHandler.StoreFile[:strings.LastIndex(s.FileHandler.StoreFile, "/")], 0777); err != nil {
			loggers.ErrorLogger.Println("failed to create directory:", err)
		}
	}
	if s.FileHandler.Restore {
		s.RestoreMetricsFromFile()
	}
	loggers.InfoLogger.Printf("Server started at %s", s.Addr)
	go repeating.Repeat(s.StoreMetricsToFile, s.FileHandler.StoreInterval)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		loggers.ErrorLogger.Fatal(err)
	}
}

func main() {
	StartServer()
}
