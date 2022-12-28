package main

import (
	"log"
	"net/http"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server"
)

const (
	Server = "127.0.0.1"
	Port   = "8080"
)

func StartServer() {
	srv := &http.Server{
		Addr: Server + ":" + Port,
	}
	http.HandleFunc("/update/", server.PostMetricHandler)
	log.Println("Server started")
	log.Fatal(srv.ListenAndServe())

}

func main() {
	StartServer()
}
