package server

import (
	"compress/gzip"
	"database/sql"

	//"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
	filestorage "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/fileStorage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/storage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const contentTypeJSON = "application/json"

type Server struct {
	Addr        string
	Debug       bool
	Key         string
	Storage     storage.Storage
	StorageType types.StorageType
}

func NewServer(address string, debug bool, fs filestorage.FileStorage, db *sql.DB, key string) *Server {
	var (
		storage     storage.Storage
		storageType types.StorageType
	)
	if db == nil {
		storage = fs
		storageType = types.StorageTypeFile
	} else {
		storage = database.NewDatabase(db)
		storageType = types.StorageTypeDB
	}
	return &Server{
		Addr:        address,
		Debug:       debug,
		Key:         key,
		Storage:     storage,
		StorageType: storageType,
	}
}

func CompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(types.GZIPWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func DecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			rw.Header().Set("Accept-Encoding", "gzip")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
			next.ServeHTTP(rw, r)
		} else {
			next.ServeHTTP(rw, r)
		}
	})

}

func GetGaugeStatusOK(rw http.ResponseWriter, metricVal float64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
	_, err := rw.Write([]byte(strVal))
	if err != nil {
		loggers.ErrorLogger.Println("response writer error:", err)
		return
	}
}

func GetCounterStatusOK(rw http.ResponseWriter, metricVal int64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	_, err := rw.Write([]byte(fmt.Sprintf("%d", metricVal)))
	if err != nil {
		loggers.ErrorLogger.Println("response writer error:", err)
		return
	}

}

func (s *Server) GetAllMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	loggers.InfoLogger.Println("Get all request")
	rw.Header().Set("Content-Type", "text/html")
	metrics, err := s.Storage.GetAllMetrics()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		loggers.ErrorLogger.Println("error while getting all metrics:", err)
		return
	}
	for _, m := range metrics {
		switch m.MType {
		case "counter":
			_, err := rw.Write([]byte(fmt.Sprintf("%s: %d", m.ID, *m.Delta)))
			if err != nil {
				http.Error(rw, fmt.Sprintf("error while writing response body: %v", err), http.StatusInternalServerError)
				loggers.ErrorLogger.Println("error while writing response body:", err)
				return
			}
		case "gauge":
			_, err := rw.Write([]byte(fmt.Sprintf("%s: %f", m.ID, *m.Value)))
			if err != nil {
				http.Error(rw, fmt.Sprintf("error while writing response body: %v", err), http.StatusInternalServerError)
				loggers.ErrorLogger.Println("error while writing response body:", err)
				return
			}
		}
		_, err := rw.Write([]byte("\n"))
		if err != nil {
			http.Error(rw, fmt.Sprintf("error while writing response body: %v", err), http.StatusInternalServerError)
			loggers.ErrorLogger.Println("error while writing response body:", err)
			return
		}
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) PostUpdateManyMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	var metrics []types.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		loggers.ErrorLogger.Println("update many decode error:", err)
	}
	if s.Debug {
		loggers.DebugLogger.Println("POST many metrics request")
	}
	err := s.Storage.SaveManyMetrics(metrics, s.Key)
	if errors.Is(err, myerrors.ErrTypeNotImplemented) {
		http.Error(rw, err.Error(), http.StatusNotImplemented)
	}
	if errors.Is(err, myerrors.ErrTypeBadRequest) {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	if err != nil {
		loggers.ErrorLogger.Println("store many metrics error:", err)
		return
	}
}

func (s *Server) PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var m types.Metrics
	metricType, metricName, metricValue := chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value")
	m.ID = metricName
	m.MType = metricType
	if m.MType == "counter" {
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		m.Delta = &delta
	}
	if m.MType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		m.Value = &value
	}
	err := s.Storage.SaveMetric(m, s.Key)
	if errors.Is(err, myerrors.ErrTypeNotImplemented) {
		http.Error(rw, err.Error(), http.StatusNotImplemented)
	}
	if errors.Is(err, myerrors.ErrTypeBadRequest) {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var m = types.Metrics{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "type"),
	}
	if s.Debug {
		loggers.DebugLogger.Printf("GET %s %s", m.MType, m.ID)
	}
	if m, err := s.Storage.GetMetric(m, s.Key); err == nil {
		if s.Debug {
			loggers.DebugLogger.Println(m.ID, *m.Delta)
		}
		switch m.MType {
		case "counter":
			if s.Debug {
				loggers.DebugLogger.Println(m.ID, *m.Delta)
			}
			GetCounterStatusOK(rw, *m.Delta)
		case "gauge":
			GetGaugeStatusOK(rw, *m.Value)
		}
	} else {
		if s.Debug {
			loggers.DebugLogger.Println("metric not found")
		}
		if errors.Is(err, myerrors.ErrTypeNotFound) {
			http.Error(rw, err.Error(), http.StatusNotFound)
		}
		if errors.Is(err, myerrors.ErrTypeBadRequest) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		if errors.Is(err, myerrors.ErrTypeNotImplemented) {
			http.Error(rw, err.Error(), http.StatusNotImplemented)
		}
	}
}

func (s *Server) PostMetricJSONHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
			return
		}
		loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
		return
	}
	var m types.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		loggers.ErrorLogger.Printf("Decode error: %v", err)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			return
		}
		return
	}
	if s.Debug {
		loggers.DebugLogger.Println("POST JSON " + m.ID + " " + m.MType)
	}
	err := s.Storage.SaveMetric(m, s.Key)
	if errors.Is(err, myerrors.ErrTypeNotImplemented) {
		http.Error(rw, err.Error(), http.StatusNotImplemented)
		return
	}
	if errors.Is(err, myerrors.ErrTypeBadRequest) {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonMetric, err := json.Marshal(m)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		loggers.ErrorLogger.Printf("json Marshal error: %s", err)
		return
	}
	_, err = rw.Write(jsonMetric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		http.Error(rw, "can't write body", http.StatusInternalServerError)
		loggers.ErrorLogger.Println(err)
		return
	}
	rw.Header().Add("Content-Type", contentTypeJSON)
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricPostJSONHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(`{"Status":"Bad Request"}`))
		if err != nil {
			loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
			return
		}
		return
	}
	body, err := io.ReadAll(r.Body)
	loggers.DebugLogger.Println(string(body))
	rw.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(rw, "reading body error", http.StatusInternalServerError)
		return
	}
	var m types.Metrics
	if err = json.Unmarshal(body, &m); err != nil {
		http.Error(rw, "Could not unmarshal JSON:", http.StatusInternalServerError)
		return
	}
	if s.Debug {
		loggers.DebugLogger.Println("Get JSON:", m)
	}
	m, err = s.Storage.GetMetric(m, s.Key)
	if s.Debug {
		loggers.DebugLogger.Println(m)
	}
	if err != nil {
		if errors.Is(err, myerrors.ErrTypeNotFound) {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		} else if errors.Is(err, myerrors.ErrTypeNotImplemented) {
			http.Error(rw, fmt.Sprintf("wrong type of metric: %v", err), http.StatusNotImplemented)
		} else if errors.Is(err, myerrors.ErrTypeBadRequest) {
			http.Error(rw, fmt.Sprintf("bad request: %v", err), http.StatusBadRequest)
		} else {
			http.Error(rw, "There is no metric you requested", http.StatusInternalServerError)
		}
		return
	}
	jsonMetric, err := json.Marshal(m)
	if err != nil {
		loggers.ErrorLogger.Printf("json Marshal error: %s", err)
		return
	}
	_, err = rw.Write(jsonMetric)
	if err != nil {
		http.Error(rw, "can't write body", http.StatusInternalServerError)
		loggers.ErrorLogger.Println(err)
		return
	}
	rw.Header().Add("Content-Type", contentTypeJSON)
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetPingDBHandler(rw http.ResponseWriter, r *http.Request) {
	if s.StorageType != types.StorageTypeDB {
		http.Error(rw, "nil database pointer", http.StatusInternalServerError)
		return
	}
	if err := s.Storage.Check(); err != nil {
		http.Error(rw, "error occured while connecting to database", http.StatusInternalServerError)
		loggers.ErrorLogger.Println("db.Ping error:", err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}
