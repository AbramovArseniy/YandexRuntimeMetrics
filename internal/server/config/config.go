package config

import (
	"database/sql"
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// defaultAddress is a default server address
const defaultAddress = "localhost:8080"

// default file storage config
const (
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
)

type Config struct {
	Address         string `json:"address"`
	Debug           bool   `json:"debug"`
	DatabaseAddress string `json:"database_dsn"`
	Database        *sql.DB
	StoreFile       string        `json:"store_file"`
	StoreInterval   time.Duration `json:"store_interval"`
	Restore         bool          `json:"restore"`
	HashKey         string
	CryptoKeyFile   string `json:"crypto_key"`
}

// SetServerParams sets server config
func SetServerParams() (cfg Config) {
	var (
		flagRestore       bool
		flagStoreFile     string
		flagAddress       string
		flagStoreInterval time.Duration
		flagDebug         bool
		flagKey           string
		flagDataBase      string
		flagCryptoKeyFile string
		flagConfig        string
	)
	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore_true/false")
	flag.StringVar(&flagStoreFile, "f", defaultStoreFile, "store_file")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.DurationVar(&flagStoreInterval, "i", defaultStoreInterval, "store_interval_in_seconds")
	flag.BoolVar(&flagDebug, "debug", false, "debug_true/false")
	flag.StringVar(&flagKey, "k", "", "hash_key")
	flag.StringVar(&flagDataBase, "d", "", "db_address")
	flag.StringVar(&flagCryptoKeyFile, "crypto-key", "", "crypto_key_file")
	flag.StringVar(&flagConfig, "c", "", "config_as_json")
	flag.Parse()
	var exists bool
	if flagConfig != "" {
		err := json.Unmarshal([]byte(flagConfig), &cfg)
		if err != nil {
			loggers.ErrorLogger.Println("error while unmarshalling config json:", err)
		}
		return cfg
	} else if config, exists := os.LookupEnv("CONFIG"); exists {
		err := json.Unmarshal([]byte(config), &cfg)
		if err != nil {
			loggers.ErrorLogger.Println("error while unmarshalling config json:", err)
		}
		return cfg
	}
	cfg.Address, exists = os.LookupEnv("ADDRESS")
	if !exists {
		cfg.Address = flagAddress
	}
	if cfg.StoreFile, exists = os.LookupEnv("STORE_FILE"); !exists {
		cfg.StoreFile = flagStoreFile
	}
	var strStoreInterval, strRestore string
	if strStoreInterval, exists = os.LookupEnv("STORE_INTERVAL"); !exists {
		cfg.StoreInterval = flagStoreInterval
	} else {
		var err error
		if cfg.StoreInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			loggers.ErrorLogger.Println("couldn't parse store interval")
			cfg.StoreInterval = flagStoreInterval
		}
	}
	if strRestore, exists = os.LookupEnv("RESTORE"); !exists {
		cfg.Restore = flagRestore
	} else {
		var err error
		if cfg.Restore, err = strconv.ParseBool(strRestore); err != nil {
			loggers.ErrorLogger.Println("couldn't parse restore bool")
			cfg.Restore = flagRestore
		}
	}
	cfg.HashKey, exists = os.LookupEnv("KEY")
	if !exists {
		cfg.HashKey = flagKey
	}
	cfg.DatabaseAddress, exists = os.LookupEnv("DATABASE_DSN")
	if !exists {
		cfg.DatabaseAddress = flagDataBase
	}
	cfg.CryptoKeyFile, exists = os.LookupEnv("CRYPTO_KEY")
	if !exists {
		cfg.CryptoKeyFile = flagCryptoKeyFile
	}
	return cfg
}
