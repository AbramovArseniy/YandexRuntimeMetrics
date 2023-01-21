package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) Router() chi.Router {
	router := chi.NewRouter()
	router.Get("/", s.handler.GetAllMetricsHandler)
	router.Get("/value/{type}/{name}", s.handler.GetMetricHandler)
	router.Post("/update/{type}/{name}/{value}", s.handler.PostMetricHandler)
	router.Post("/update/", s.handler.PostMetricJSONHandler)
	router.Post("/value/", s.handler.GetMetricPostJSONHandler)
	return router
}
