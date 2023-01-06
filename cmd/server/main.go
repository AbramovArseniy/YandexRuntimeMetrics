package main

import (
	"log"
	"net/http"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
)

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = "8080"
)

func StartServer() {
	srv := &http.Server{
		Addr:    DefaultHost + ":" + DefaultPort,
		Handler: server.Router(),
	}
	log.Println("Server started")
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func main() {
	StartServer()
}
