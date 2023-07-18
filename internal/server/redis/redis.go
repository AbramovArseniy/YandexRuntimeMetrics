package redis

import (
	"crypto/hmac"
	"errors"
	"fmt"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/hash"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
	"github.com/gomodule/redigo/redis"
)

type Database struct {
	Conn redis.Conn
}

func NewDatabase(address string) (*Database, error) {
	conn, err := redis.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}
	return &Database{
		Conn: conn,
	}, nil
}

func (db *Database) GetAllMetrics() ([]types.Metrics, error) {
	metrics := make([]types.Metrics, 0)
	keys, err := redis.Strings(db.Conn.Do("KEYS", "*"))
	if err != nil {
		loggers.ErrorLogger.Println("db query error:", err)
		return nil, err
	}
	for _, key := range keys {
		response, err := redis.Values(db.Conn.Do("HGET", key))
		if err != nil {
			loggers.ErrorLogger.Println("db query error:", err)
			return nil, err
		}
		var m types.Metrics
		err = redis.ScanStruct(response, &m)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// GetMetrics gets info about one metric
func (db *Database) GetMetric(m types.Metrics, key string) (types.Metrics, error) {
	switch m.MType {
	case "gauge":
		value, err := redis.Float64(db.Conn.Do("HGET", m.MType+":"+m.ID, "value"))
		if errors.Is(err, redis.ErrNil) {
			return m, myerrors.ErrTypeNotFound
		}
		if err != nil {
			loggers.ErrorLogger.Println("db query error:", err)
			return m, err
		}
		m.Value = &value
		if key != "" {
			metricHash := hash.Hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key)
			m.Hash = string(metricHash)
		}
	case "counter":
		delta, err := redis.Int64(db.Conn.Do("HGET", m.MType+":"+m.ID, "delta"))
		if errors.Is(err, redis.ErrNil) {
			return m, myerrors.ErrTypeNotFound
		}
		if err != nil {
			loggers.ErrorLogger.Println("db query error:", err)
			return m, err
		}
		m.Delta = &delta
		if key != "" {
			metricHash := hash.Hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key)
			m.Hash = string(metricHash)
		}
	default:
		return m, myerrors.ErrTypeNotFound
	}
	return m, nil
}

// SaveMetric saves info about one metric
func (db Database) SaveMetric(m types.Metrics, key string) error {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return fmt.Errorf("%wno value in update request", myerrors.ErrTypeBadRequest)
		}
		if key != "" && m.Hash != "" {
			if !hmac.Equal([]byte(m.Hash), []byte(hash.Hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key))) {
				return fmt.Errorf("%wwrong hash in request", myerrors.ErrTypeBadRequest)
			}
		}
		_, err := db.Conn.Do("HMSET", m.MType+":"+m.ID, "id", m.ID, "type", m.MType, "delta", nil, "value", *m.Value, "hash", m.Hash)
		if err != nil {
			return err
		}
	case "counter":
		if m.Delta == nil {
			return fmt.Errorf("%wno value in update request", myerrors.ErrTypeBadRequest)
		}
		if key != "" && m.Hash != "" {
			if !hmac.Equal([]byte(m.Hash), []byte(hash.Hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key))) {
				return fmt.Errorf("%wwrong hash in request", myerrors.ErrTypeBadRequest)
			}
		}
		metric, err := db.GetMetric(m, key)
		if errors.Is(err, myerrors.ErrTypeNotFound) {
			_, err := db.Conn.Do("HMSET", m.MType+":"+m.ID, "id", m.ID, "type", m.MType, "delta", *m.Delta, "value", nil, "hash", m.Hash)
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}
		_, err = db.Conn.Do("HMSET", m.MType+":"+m.ID, "id", m.ID, "type", m.MType, "delta", *metric.Delta+*m.Delta, "value", nil, "hash", m.Hash)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%wno such type of metric", myerrors.ErrTypeNotImplemented)
	}
	return nil
}

// SaveManyMetrics saves info about several metrics
func (db *Database) SaveManyMetrics(metrics []types.Metrics, key string) error {
	for _, m := range metrics {
		err := db.SaveMetric(m, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Check checks if storage works OK
func (db *Database) Check() error {
	return nil
}
