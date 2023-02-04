package server

import (
	"database/sql"
	"io"
	"net/http"
	"time"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type fileHandler struct {
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

type Server struct {
	Addr                              string
	storage                           MemStorage
	FileHandler                       fileHandler
	Debug                             bool
	Key                               string
	DataBase                          *sql.DB
	InsertUpdateCounterToDatabaseStmt *sql.Stmt
	InsertUpdateGaugeToDatabaseStmt   *sql.Stmt
	SelectAllFromDatabaseStmt         *sql.Stmt
	SelectOneGaugeFromDatabaseStmt    *sql.Stmt
	SelectOneCounterFromDatabaseStmt  *sql.Stmt
}

func NewServer(address string, storeInterval time.Duration, storeFile string, restore bool, debug bool, key string, db *sql.DB) *Server {
	var insertCounterStmt, insertGaugeStmt, selectAllStmt, selectOneGaugeStmt, selectOneCounterStmt *sql.Stmt = nil, nil, nil, nil, nil
	if db != nil {
		var err error
		insertCounterStmt, err = db.Prepare(`
			INSERT INTO metrics (id, type, value, delta) VALUES ($1, 'counter', NULL, $2)
			ON CONFLICT (id, type) DO UPDATE SET
				value=EXCLUDED.value,
				delta=EXCLUDED.delta;
		`)
		if err != nil {
			loggers.ErrorLogger.Println("insert counter statement prepare error:", err)
		}
		insertGaugeStmt, err = db.Prepare(`
			INSERT INTO metrics (id, type, value, delta) VALUES ($1, 'gauge', $2, NULL)
			ON CONFLICT (id, type) DO UPDATE SET
				value=EXCLUDED.value,
				delta=EXCLUDED.delta;
		`)
		if err != nil {
			loggers.ErrorLogger.Println("insert statement prepare error:", err)
		}
		selectAllStmt, err = db.Prepare(`SELECT id, type, value, delta FROM metrics;`)
		if err != nil {
			loggers.ErrorLogger.Println("select all statement prepare error:", err)
		}
		selectOneGaugeStmt, err = db.Prepare(`SELECT value FROM metrics WHERE id=$1;`)
		if err != nil {
			loggers.ErrorLogger.Println("select one statement prepare error:", err)
		}
		selectOneCounterStmt, err = db.Prepare(`SELECT value FROM metrics WHERE id=$1;`)
		if err != nil {
			loggers.ErrorLogger.Println("select one statement prepare error:", err)
		}
	}
	return &Server{
		Addr: address,
		storage: MemStorage{
			CounterMetrics: make(map[string]int64),
			GaugeMetrics:   make(map[string]float64),
		},
		FileHandler: fileHandler{
			StoreInterval: storeInterval,
			StoreFile:     storeFile,
			Restore:       restore,
		},
		Debug:                             debug,
		Key:                               key,
		DataBase:                          db,
		InsertUpdateCounterToDatabaseStmt: insertCounterStmt,
		InsertUpdateGaugeToDatabaseStmt:   insertGaugeStmt,
		SelectAllFromDatabaseStmt:         selectAllStmt,
		SelectOneGaugeFromDatabaseStmt:    selectOneGaugeStmt,
		SelectOneCounterFromDatabaseStmt:  selectOneCounterStmt,
	}
}
