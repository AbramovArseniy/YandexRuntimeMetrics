package server

import (
	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	h := Handler{
		storage: MemStorage{
			CounterMetrics: make(map[string]int64),
			GaugeMetrics:   make(map[string]float64),
		},
	}
	router := chi.NewRouter()
	router.Get("/", h.GetAllMetricsHandler)
	router.Get("/value/{type}/{name}", h.GetMetricHandler)
	router.Post("/update/{type}/{name}/{value}", h.PostMetricHandler)
	router.Post("/update", h.PostMetricJsonHandler)
	router.Post("/value", h.GetMetricPostJsonHandler)
	return router
}
