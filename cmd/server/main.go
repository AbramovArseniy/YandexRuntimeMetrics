// Package main starts server
package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	pb "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/proto"
	serverHTTP "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/HTTP"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
	servergRPC "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/gRPC"
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
	if cfg.Protocol == "HTTP" {
		s := serverHTTP.NewMetricServer(cfg)
		handler := serverHTTP.DecompressHandler(s.Router())
		handler = serverHTTP.CompressHandler(handler)
		srv := &http.Server{
			Addr:    s.Addr,
			Handler: handler,
		}
		loggers.InfoLogger.Printf("HTTP server started at %s", s.Addr)
		idleConnsClosed := make(chan struct{})
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		go func() {
			<-sigs
			if err := srv.Shutdown(context.Background()); err != nil {
				loggers.InfoLogger.Printf("HTTP server Shutdown: %v", err)
			}
			close(idleConnsClosed)
		}()
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			loggers.ErrorLogger.Fatal(err)
		}
	} else if cfg.Protocol == "gRPC" {
		s := servergRPC.NewMetricServer(cfg)
		listen, err := net.Listen("tcp", s.Addr)
		if err != nil {
			loggers.ErrorLogger.Fatal(err)
		}
		srv := grpc.NewServer(grpc.UnaryInterceptor(s.CheckRequestSubnetInterceptor))
		pb.RegisterMetricsServer(srv, s)
		loggers.InfoLogger.Println("gRPC server started at", s.Addr)
		if err := srv.Serve(listen); err != nil {
			loggers.ErrorLogger.Fatal(err)
		}
	}
	loggers.InfoLogger.Printf(`Build version: %s
	Build date: %s
	Build commit: %s`,
		buildVersion, buildDate, buildCommit)

}

func main() {
	StartServer()
}
