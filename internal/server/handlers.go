package server

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
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
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const contentTypeJSON = "application/json"

var ErrTypeNotImplemented = errors.New("not implemented: ")
var ErrTypeBadRequest = errors.New("bad request: ")
var ErrTypeNotFound = errors.New("not found: ")

func hash(src, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
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
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
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
	if s.Database.DB == nil {
		for metricName, metricVal := range s.storage.GaugeMetrics {
			strVal := strconv.FormatFloat(metricVal, 'f', -1, 64)
			_, err := rw.Write([]byte(fmt.Sprintf("%s: %s\n", metricName, strVal)))
			if err != nil {
				loggers.ErrorLogger.Println("response writer error:", err)
				return
			}
		}
		for metricName, metricVal := range s.storage.CounterMetrics {
			_, err := rw.Write([]byte(fmt.Sprintf("%s: %d", metricName, metricVal)))
			if err != nil {
				loggers.ErrorLogger.Println("response writer error:", err)
				return
			}
		}
	} else {
		rows, err := s.Database.DB.QueryContext(r.Context(), "SELECT metrics.id, type, value, delta FROM metrics")
		if err != nil {
			http.Error(rw, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			var (
				id, mType string
				delta     sql.NullInt64
				value     sql.NullFloat64
			)
			rows.Scan(&id, &mType, &delta, &value)
			switch mType {
			case "counter":
				_, err := rw.Write([]byte(fmt.Sprintf("%s: %d", id, delta.Int64)))
				if err != nil {
					loggers.ErrorLogger.Println("response writer error:", err)
					return
				}
			case "gauge":
				_, err := rw.Write([]byte(fmt.Sprintf("%s: %f", id, value.Float64)))
				if err != nil {
					loggers.ErrorLogger.Println("response writer error:", err)
					return
				}
			}
			if rows.Err() != nil {
				http.Error(rw, fmt.Sprintf("database error: %v", rows.Err()), http.StatusInternalServerError)
				return
			}
		}
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) PostUpdateManyMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	var metrics []Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		loggers.ErrorLogger.Println("update many decode error:", err)
	}
	if s.Debug {
		loggers.DebugLogger.Println("POST many metrics request")
	}
	err := s.StoreManyMetrics(metrics)
	if err != nil {
		loggers.ErrorLogger.Println("store many metrics error:", err)
		return
	}
}

func (s *Server) PostMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType, metricName, metricValue := chi.URLParam(r, "type"), chi.URLParam(r, "name"), chi.URLParam(r, "value")
	switch metricType {
	case "gauge":
		newVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, fmt.Sprintf("error parsing value %s as float", metricValue), http.StatusBadRequest)
		}
		s.storage.GaugeMetrics[metricName] = newVal
	case "counter":
		newVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, fmt.Sprintf("error parsing value %s as int", metricValue), http.StatusBadRequest)
		}
		s.storage.CounterMetrics[metricName] += newVal
	default:
		loggers.ErrorLogger.Printf("wrong Metric Type: %s", metricType)
		http.Error(rw, "Wrong Metric Type", http.StatusNotImplemented)
	}
	if s.Debug {
		loggers.DebugLogger.Printf("POST %s %s", metricType, metricName)
	}
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var m = Metrics{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "type"),
	}
	if s.Debug {
		loggers.DebugLogger.Printf("GET %s %s", m.MType, m.ID)
	}
	if m, err := s.GetMetricValue(rw, r, m); err == nil {
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
		if errors.Is(err, ErrTypeNotFound) {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
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
	var m Metrics
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
	if s.Database.DB != nil {
		err := s.storeMetricsToDatabase(m)
		if err != nil {
			if errors.Is(err, ErrTypeNotImplemented) {
				http.Error(rw, err.Error(), http.StatusNotImplemented)
			}
			loggers.ErrorLogger.Println("store metrics to db error:", err.Error())
			return
		}
	} else {
		err := s.storeMetrics(m)
		if err != nil {
			if errors.Is(err, ErrTypeNotImplemented) {
				http.Error(rw, err.Error(), http.StatusNotImplemented)
			}
			if errors.Is(err, ErrTypeBadRequest) {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			loggers.ErrorLogger.Println("Store Metrics error:", err.Error())
			return
		}
	}
	rw.Header().Add("Content-Type", contentTypeJSON)
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
	var m Metrics
	if err = json.Unmarshal(body, &m); err != nil {
		http.Error(rw, "Could not unmarshal JSON:", http.StatusInternalServerError)
		return
	}
	if s.Debug {
		loggers.DebugLogger.Println("Get JSON:", m)
	}
	m, err = s.GetMetricValue(rw, r, m)
	if s.Debug {
		loggers.DebugLogger.Println(m)
	}
	if err != nil {
		if errors.Is(err, ErrTypeNotFound) {
			http.Error(rw, "There is no metric you requested", http.StatusNotFound)
		} else {
			http.Error(rw, fmt.Sprintf("Err while getting value: %v", err), http.StatusInternalServerError)
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

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetPingDBHandler(rw http.ResponseWriter, r *http.Request) {
	if s.Database.DB == nil {
		http.Error(rw, "nil database pointer", http.StatusInternalServerError)
		return
	}
	if err := s.Database.DB.Ping(); err != nil {
		http.Error(rw, "error occured while connecting to database", http.StatusInternalServerError)
		loggers.ErrorLogger.Println("db.Ping error:", err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricValue(rw http.ResponseWriter, r *http.Request, m Metrics) (Metrics, error) {
	if s.Database.DB != nil {
		switch m.MType {
		case "gauge":
			var value float64
			err := s.Database.SelectOneGaugeFromDatabaseStmt.QueryRowContext(r.Context(), m.ID).Scan(&value)
			if err != nil {
				loggers.ErrorLogger.Println("db query error:", err)
				return m, err
			}
			m.Value = &value
			if s.Key != "" {
				metricHash := hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key)
				m.Hash = string(metricHash)
			}
		case "counter":
			var delta int64
			err := s.Database.SelectOneCounterFromDatabaseStmt.QueryRowContext(r.Context(), m.ID).Scan(&delta)
			if err != nil {
				return m, ErrTypeNotFound
			}
			m.Delta = &delta
			if s.Key != "" {
				metricHash := hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key)
				m.Hash = string(metricHash)
			}
		default:
			if s.Debug {
				loggers.DebugLogger.Println("There is no metric you requested")
			}
			return m, ErrTypeNotFound
		}
	} else {
		switch m.MType {
		case "counter":
			val, isIn := s.storage.CounterMetrics[m.ID]
			if !isIn {
				if s.Debug {
					loggers.DebugLogger.Println("There is no metric you requested")
				}
				return m, ErrTypeNotFound
			}
			m.Delta = &val
			if s.Key != "" {
				metricHash := hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key)
				m.Hash = string(metricHash)
			}
		case "gauge":
			val, isIn := s.storage.GaugeMetrics[m.ID]
			if !isIn {
				if s.Debug {
					loggers.DebugLogger.Println("There is no metric you requested")
				}
				return m, ErrTypeNotFound
			}
			m.Value = &val
			if s.Key != "" {
				metricHash := hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key)
				m.Hash = string(metricHash)
			}
		default:
			if s.Debug {
				loggers.DebugLogger.Println("There is no metric you requested")
			}
			return m, ErrTypeNotFound
		}
	}
	return m, nil
}

func (s *Server) storeMetrics(m Metrics) error {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		if s.Key != "" && m.Hash != "" {
			if s.Debug {
				loggers.DebugLogger.Println(m.Hash, "     -     ", hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key))
			}
			if !hmac.Equal([]byte(m.Hash), []byte(hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key))) {
				return fmt.Errorf("%wwrong hash in request", ErrTypeBadRequest)
			}
		}
		s.storage.GaugeMetrics[m.ID] = *m.Value
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		if s.Key != "" && m.Hash != "" {
			if s.Debug {
				loggers.DebugLogger.Println(m.Hash, "     -     ", hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key))
			}
			if !hmac.Equal([]byte(m.Hash), []byte(hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key))) {
				return fmt.Errorf("%wwrong hash in request", ErrTypeBadRequest)
			}
		}
		s.storage.CounterMetrics[m.ID] += *m.Delta
	default:
		return fmt.Errorf("%wno such type of metric", ErrTypeNotImplemented)
	}
	return nil
}

func (s *Server) storeMetricsToDatabase(m Metrics) error {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		if s.Key != "" && m.Hash != "" {
			if s.Debug {
				loggers.DebugLogger.Println(m.Hash, "     -     ", hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key))
			}
			if !hmac.Equal([]byte(m.Hash), []byte(hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), s.Key))) {
				return fmt.Errorf("%wwrong hash in request", ErrTypeBadRequest)
			}
		}
		_, err := s.Database.InsertUpdateGaugeToDatabaseStmt.Exec(m.ID, *m.Value)
		if err != nil {
			return err
		}
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("%wno value in update request", ErrTypeNotImplemented)
		}
		if s.Key != "" && m.Hash != "" {
			if s.Debug {
				loggers.DebugLogger.Println(m.Hash, "     -     ", hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key))
			}
			if !hmac.Equal([]byte(m.Hash), []byte(hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), s.Key))) {
				return fmt.Errorf("%wwrong hash in request", ErrTypeBadRequest)
			}
		}
		var numberOfMetrics int
		err := s.Database.CountIDsInDatabaseStmt.QueryRow(m.ID).Scan(&numberOfMetrics)
		if err != nil {
			return err
		}
		if numberOfMetrics != 0 {
			var delta int64
			err = s.Database.SelectOneCounterFromDatabaseStmt.QueryRow(m.ID).Scan(&delta)
			if err != nil {
				return err
			}
			_, err = s.Database.UpdateCounterToDatabaseStmt.Exec(m.ID, delta+*m.Delta)
			if err != nil {
				return err
			}
		} else {
			_, err = s.Database.InsertCounterToDatabaseStmt.Exec(m.ID, *m.Delta)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%wno such type of metric", ErrTypeNotImplemented)
	}
	return nil
}

func (s *Server) StoreManyMetrics(metrics []Metrics) error {
	var err error
	for _, m := range metrics {
		if s.Database.DB != nil {
			err = s.storeMetricsToDatabase(m)
		} else {
			err = s.storeMetrics(m)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
