package server

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "net/http/pprof"
	"os"

	//"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
	filestorage "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/fileStorage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/storage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
)

const contentTypeJSON = "application/json"

// Server has server info
type Server struct {
	Addr        string
	Debug       bool
	Key         string
	Storage     storage.Storage
	StorageType types.StorageType
	CryptoKey   *rsa.PrivateKey
}

// NewServer creates new Server
func NewServer(cfg config.Config) *Server {
	var (
		storage     storage.Storage
		storageType types.StorageType
	)

	var cryptoKey *rsa.PrivateKey
	if cfg.CryptoKeyFile != "" {
		file, err := os.OpenFile(cfg.CryptoKeyFile, os.O_RDONLY, 0777)
		if err != nil {
			loggers.ErrorLogger.Println("error while opening crypto key file:", err)
			cryptoKey = nil
		} else {
			cryptoKeyByte, err := io.ReadAll(file)
			if err != nil {
				loggers.ErrorLogger.Println("error while reading crypto key file:", err)
				cryptoKey = nil
			}
			cryptoKey, err = x509.ParsePKCS1PrivateKey(cryptoKeyByte)
			if err != nil {
				loggers.ErrorLogger.Println("error while parsing crypto key:", err)
				cryptoKey = nil
			}
		}
	}
	if cfg.Database == nil {
		fs := filestorage.NewFileStorage(cfg)
		fs.SetFileStorage()
		storage = fs
		storageType = types.StorageTypeFile
	} else {
		storage = database.NewDatabase(cfg.Database)
		storageType = types.StorageTypeDB
	}
	return &Server{
		Addr:        cfg.Address,
		Debug:       cfg.Debug,
		Key:         cfg.HashKey,
		Storage:     storage,
		StorageType: storageType,
		CryptoKey:   cryptoKey,
	}
}

// CompressHandler is a middleware that compresses data to gzip if gzip encoding is accepted
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

func (s *Server) DecodeHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			loggers.ErrorLogger.Println("DecodeHandler: error while reading request body:", err)
		} else {
			decodedBody, err := rsa.DecryptPKCS1v15(rand.Reader, s.CryptoKey, body)
			if err != nil {
				loggers.ErrorLogger.Println("DecodeHandler: error while reading request body:", err)
			} else {
				newReqBody := bytes.NewReader(decodedBody)
				r.Body = io.NopCloser(newReqBody)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// DecompressHandler is a middleware that decompresses data from gzip
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

// GetGaugeStatusOK describes response in case of successful getting of gauge metric value from storage
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

// GetCounterStatusOK describes response in case of successful getting of counter metric value from storage
func GetCounterStatusOK(rw http.ResponseWriter, metricVal int64) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain")
	_, err := fmt.Fprintf(rw, "%d", metricVal)
	if err != nil {
		loggers.ErrorLogger.Println("response writer error:", err)
		return
	}

}

// GetAllMetricsHandler prints info about all metrics in storage
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
			_, err := fmt.Fprintf(rw, "%s: %d", m.ID, *m.Delta)
			if err != nil {
				http.Error(rw, fmt.Sprintf("error while writing response body: %v", err), http.StatusInternalServerError)
				loggers.ErrorLogger.Println("error while writing response body:", err)
				return
			}
		case "gauge":
			_, err := fmt.Fprintf(rw, "%s: %f", m.ID, *m.Value)
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

// PostUpdateManyMetricsHandler updates info about several metrics
func (s *Server) PostUpdateManyMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		http.Error(rw, "wrong content type", http.StatusBadRequest)
		loggers.ErrorLogger.Println("Wrong content type:", r.Header.Get("Content-Type"))
		return
	}
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
	byteResponse, err := json.Marshal(metrics)
	if err != nil {
		loggers.ErrorLogger.Println("error while marshaling many metrics update response:", err)
	}
	rw.Write(byteResponse)
	rw.Header().Add("Content-Type", contentTypeJSON)
}

// PostMetricHandler updates info about one metric
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

// GetMetricHandler prints value of requested metric
func (s *Server) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var m = types.Metrics{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "type"),
	}
	var err error
	if s.Debug {
		loggers.DebugLogger.Printf("GET %s %s", m.MType, m.ID)
	}
	if m, err = s.Storage.GetMetric(m, s.Key); err == nil {
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

// PostMetricJSONHandler updates info about one metric, sent a json
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

// GetMetricPostJSONHandler prints info about metrics requested as json
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

// GetPingDBHandler checks if database is connected
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
