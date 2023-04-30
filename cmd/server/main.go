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
	var err error
	if cfg.DatabaseAddress != "" {
		cfg.Database, err = sql.Open("pgx", cfg.DatabaseAddress)
		if err != nil {
			loggers.ErrorLogger.Println("opening DB error:", err)
			cfg.Database = nil
		} else {
			err = database.SetDatabase(cfg.Database, cfg.DatabaseAddress)
			if err != nil {
				loggers.ErrorLogger.Println("error while setting database:", err)
			}
		}
		defer cfg.Database.Close()
	} else {
		cfg.Database = nil
	}
	s := server.NewServer(cfg)
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
