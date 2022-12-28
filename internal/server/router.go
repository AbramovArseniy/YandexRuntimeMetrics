package server

import (
	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	router := chi.NewRouter()
	router.Get("/", GetAllMetricsHandler)
	router.Get("/value/{type}/{name}", GetMetricHandler)
	router.Post("/update/{type}/{name}/{value}", PostMetricHandler)
	return router
}
