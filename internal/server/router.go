package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) Router() chi.Router {
	router := chi.NewRouter()
	router.Get("/", s.GetAllMetricsHandler)
	router.Get("/value/{type}/{name}", s.GetMetricHandler)
	router.Post("/update/{type}/{name}/{value}", s.PostMetricHandler)
	router.Post("/update/", s.PostMetricJSONHandler)
	router.Post("/value/", s.GetMetricPostJSONHandler)
	router.Get("/ping", s.GetPingDBHandler)
	router.Post("/updates/", s.PostUpdateManyMetricsHandler)
	return router
}
