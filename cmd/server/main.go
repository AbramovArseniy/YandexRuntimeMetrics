package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"

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

func setServerParams() (string, time.Duration, string, bool, bool, string, string) {
	var (
		flagRestore, restore             bool
		flagStoreFile, storeFile         string
		flagAddress                      string
		flagStoreInterval, storeInterval time.Duration
		flagDebug                        bool
		flagKey                          string
		flagDataBase                     string
	)

	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore_true/false")
	flag.StringVar(&flagStoreFile, "f", defaultStoreFile, "store_file")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.DurationVar(&flagStoreInterval, "i", defaultStoreInterval, "store_interval_in_seconds")
	flag.BoolVar(&flagDebug, "debug", false, "debug_true/false")
	flag.StringVar(&flagKey, "k", "", "hash_key")
	flag.StringVar(&flagDataBase, "d", "", "db_address")
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
	key, exists := os.LookupEnv("KEY")
	if !exists {
		key = flagKey
	}
	database, exists := os.LookupEnv("DATABASE_DSN")
	if !exists {
		database = flagDataBase
	}
	return address, storeInterval, storeFile, restore, flagDebug, key, database
}

func setDatabase(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./cmd/server/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migration: %w", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func StartServer() {
	address, storeInterval, storeFile, restore, debug, key, dbAddress := setServerParams()
	var db *sql.DB
	var err error
	if dbAddress != "" {
		db, err = sql.Open("pgx", dbAddress)
		if err != nil {
			loggers.ErrorLogger.Println("opening DB error:", err)
			db = nil
		} else {
			setDatabase(db)
		}
		defer db.Close()
	} else {
		db = nil
	}
	s := server.NewServer(address, storeInterval, storeFile, restore, debug, key, db)
	handler := server.DecompressHandler(s.Router())
	handler = server.CompressHandler(handler)
	srv := &http.Server{
		Addr:    s.Addr,
		Handler: handler,
	}
	if db == nil {
		if strings.LastIndex(s.FileHandler.StoreFile, "/") != -1 {
			if err := os.MkdirAll(s.FileHandler.StoreFile[:strings.LastIndex(s.FileHandler.StoreFile, "/")], 0777); err != nil {
				loggers.ErrorLogger.Println("failed to create directory:", err)
			}
		}
		if s.FileHandler.Restore {
			s.RestoreMetricsFromFile()
		}
		go repeating.Repeat(s.StoreMetricsToFile, s.FileHandler.StoreInterval)
	}
	loggers.InfoLogger.Printf("Server started at %s", s.Addr)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		loggers.ErrorLogger.Fatal(err)
	}
}

func main() {
	StartServer()
}
