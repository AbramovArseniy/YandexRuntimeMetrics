package main

import (
	"log"
	"net/http"
	"os"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
)

const (
	defaultAddress = "localhost:8080"
)

func StartServer() {
	srv := &http.Server{
		Handler: server.Router(),
	}
	addr, exists := os.LookupEnv("ADDRESS")
	if !exists {
		srv.Addr = defaultAddress
	} else {
		srv.Addr = addr
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
