// Package main starts server
package main

import (
	"database/sql"
	"net/http"
	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
)

// build info
var buildVersion, buildDate, buildCommit string = "N/A", "N/A", "N/A"

// StartServer starts server
func StartServer() {
	cfg := config.SetServerParams()
	var db *sql.DB
	var err error
	if cfg.DatabaseAddress != "" {
		db, err = sql.Open("pgx", cfg.DatabaseAddress)
		if err != nil {
			loggers.ErrorLogger.Println("opening DB error:", err)
			db = nil
		} else {
			err = database.SetDatabase(db, cfg.DatabaseAddress)
			if err != nil {
				loggers.ErrorLogger.Println("error while setting database:", err)
			}
		}
		defer db.Close()
	} else {
		db = nil
	}
	s := server.NewServer(cfg)
	handler := server.DecompressHandler(s.Router())
	handler = server.CompressHandler(handler)
	srv := &http.Server{
		Addr:    cfg.Address,
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
