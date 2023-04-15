// Package main starts server
package main

import (
	"database/sql"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
	filestorage "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/fileStorage"
)

// defaultAddress is a default server address
const defaultAddress = "localhost:8080"

// default file storage config
const (
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
)

var buildVersion, buildDate, buildCommit string = "N/A", "N/A", "N/A"

// setServerParams sets server config
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
	var strStoreInterval, strRestore, key, database string
	if strStoreInterval, exists = os.LookupEnv("STORE_INTERVAL"); !exists {
		storeInterval = flagStoreInterval
	} else {
		var err error
		if storeInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			loggers.ErrorLogger.Println("couldn't parse store interval")
			storeInterval = flagStoreInterval
		}
	}
	if strRestore, exists = os.LookupEnv("RESTORE"); !exists {
		restore = flagRestore
	} else {
		var err error
		if restore, err = strconv.ParseBool(strRestore); err != nil {
			loggers.ErrorLogger.Println("couldn't parse restore bool")
			restore = flagRestore
		}
	}
	key, exists = os.LookupEnv("KEY")
	if !exists {
		key = flagKey
	}
	database, exists = os.LookupEnv("DATABASE_DSN")
	if !exists {
		database = flagDataBase
	}
	return address, storeInterval, storeFile, restore, flagDebug, key, database
}

// StartServer starts server
func StartServer() {
	address, storeInterval, storeFile, restore, debug, key, dbAddress := setServerParams()
	var db *sql.DB
	var fs filestorage.FileStorage
	var err error
	if dbAddress != "" {
		db, err = sql.Open("pgx", dbAddress)
		if err != nil {
			loggers.ErrorLogger.Println("opening DB error:", err)
			db = nil
		} else {
			err = database.SetDatabase(db, dbAddress)
			if err != nil {
				loggers.ErrorLogger.Println("error while setting database:", err)
			}
		}
		defer db.Close()
	} else {
		db = nil
	}
	if db == nil {
		fs = filestorage.NewFileStorage(storeFile, storeInterval, restore)
		fs.SetFileStorage()
	}
	s := server.NewServer(address, debug, fs, db, key)
	handler := server.DecompressHandler(s.Router())
	handler = server.CompressHandler(handler)
	srv := &http.Server{
		Addr:    s.Addr,
		Handler: handler,
	}

	loggers.InfoLogger.Printf("Server started at %s", s.Addr)
	loggers.InfoLogger.Printf(`Build version: %s
	Build date: %s
	Build commit: %s`,
		buildVersion, buildDate, buildCommit)
	err = http.ListenAndServe(srv.Addr, srv.Handler)
	if err != nil && err != http.ErrServerClosed {
		loggers.ErrorLogger.Fatal(err)
	}
}

func main() {
	StartServer()
}
