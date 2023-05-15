package httpserver

import (
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

// Router routes handlers to urls
func (s *MetricServer) Router() chi.Router {
	router := chi.NewRouter()
	router.Get("/", s.GetAllMetricsHandler)
	router.Get("/value/{type}/{name}", s.GetMetricHandler)
	router.Post("/update/{type}/{name}/{value}", s.PostMetricHandler)
	router.Post("/update/", s.PostMetricJSONHandler)
	router.Post("/value/", s.GetMetricPostJSONHandler)
	router.Get("/ping", s.GetPingDBHandler)
	router.Post("/updates/", s.PostUpdateManyMetricsHandler)
	router.Get("/debug/pprof/", pprof.Index)
	router.Get("/debug/pprof/cmdline", pprof.Cmdline)
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/symbol", pprof.Symbol)
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	return router
}
