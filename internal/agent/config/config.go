package config

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

// default agent preferences
const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultAddress        = "localhost:8080"
	defaultRateLimit      = 100
)

type Config struct {
	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	Address        string        `json:"address"`
	RateLimit      int
	HashKey        string
	CryptoKeyFile  string `json:"crypto_key"`
	Protocol       string
	HostAddress    string
}

// setAgentParams set agent config
func SetAgentParams() (cfg Config) {
	var (
		flagPollInterval   time.Duration
		flagReportInterval time.Duration
		flagRateLimit      int
		flagAddress        string
		flagKey            string
		flagCryptoKeyFile  string
		flagConfigFile     string
		flagProtocol       string
		cfgFile            string
	)
	flag.DurationVar(&flagPollInterval, "p", defaultPollInterval, "poll_metrics_interval")
	flag.DurationVar(&flagReportInterval, "r", defaultReportInterval, "report_metrics_interval")
	flag.IntVar(&flagRateLimit, "l", defaultRateLimit, "rate_limit")
	flag.StringVar(&flagAddress, "a", defaultAddress, "server_address")
	flag.StringVar(&flagKey, "k", "", "hash_key")
	flag.StringVar(&flagCryptoKeyFile, "crypto-key", "", "crypto_key_file")
	flag.StringVar(&flagConfigFile, "c", "", "config_file_name")
	flag.StringVar(&flagProtocol, "protocol", "HTTP", "protocol_HTTP_or_gRPC")
	flag.Parse()
	var exists bool
	if cfgFile, exists = os.LookupEnv("CONFIG"); !exists {
		cfgFile = flagConfigFile
	}
	if cfgFile != "" {
		file, err := os.Open(cfgFile)
		if err != nil {
			loggers.ErrorLogger.Println("error while opening config file:", err)
		}
		cfgJSON, err := io.ReadAll(file)
		if err != nil {
			loggers.ErrorLogger.Println("error while reading from config file:", err)
		}
		err = json.Unmarshal(cfgJSON, &cfg)
		if err != nil {
			loggers.ErrorLogger.Println("error while unmarshalling config json:", err)
		}
	}
	if strPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		cfg.PollInterval = flagPollInterval
	} else {
		var err error
		if cfg.PollInterval, err = time.ParseDuration(strPollInterval); err != nil || cfg.PollInterval <= 0 {
			log.Println("couldn't parse poll duration from environment")
			cfg.PollInterval = flagPollInterval
		}
	}
	if strReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		cfg.ReportInterval = flagReportInterval
	} else {
		var err error
		if cfg.ReportInterval, err = time.ParseDuration(strReportInterval); err != nil || cfg.ReportInterval <= 0 {
			log.Println("couldn't parse report duration from")
			cfg.ReportInterval = flagReportInterval
		}
	}
	if strRateLimit, exists := os.LookupEnv("RATE_LIMIT"); !exists {
		cfg.RateLimit = flagRateLimit
	} else {
		var err error
		if cfg.RateLimit, err = strconv.Atoi(strRateLimit); err != nil || cfg.RateLimit <= 0 {
			log.Println("couldn't parse report duration from", strRateLimit)
			cfg.RateLimit = flagRateLimit
		}
	}
	cfg.Address, exists = os.LookupEnv("ADDRESS")
	if !exists {
		cfg.Address = flagAddress
	}
	cfg.HashKey, exists = os.LookupEnv("KEY")
	if !exists {
		cfg.HashKey = flagKey
	}
	cfg.CryptoKeyFile, exists = os.LookupEnv("CRYPTO_KEY")
	if !exists {
		cfg.CryptoKeyFile = flagCryptoKeyFile
	}
	cfg.Protocol = flagProtocol
	return cfg
}
