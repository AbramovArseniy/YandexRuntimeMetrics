package database

import (
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/hash"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Database struct {
	DB                               *sql.DB
	InsertCounterToDatabaseStmt      *sql.Stmt
	UpdateCounterToDatabaseStmt      *sql.Stmt
	InsertUpdateGaugeToDatabaseStmt  *sql.Stmt
	SelectAllFromDatabaseStmt        *sql.Stmt
	SelectOneGaugeFromDatabaseStmt   *sql.Stmt
	SelectOneCounterFromDatabaseStmt *sql.Stmt
	CountIDsInDatabaseStmt           *sql.Stmt
}

func NewDatabase(db *sql.DB) Database {
	var insertCounterStmt, updateCounterStmt, countIDsStmt, insertGaugeStmt, selectAllStmt, selectOneGaugeStmt, selectOneCounterStmt *sql.Stmt = nil, nil, nil, nil, nil, nil, nil
	if db != nil {
		var err error
		countIDsStmt, err = db.Prepare("SELECT COUNT(*) FROM metrics WHERE id=$1;")
		if err != nil {
			loggers.ErrorLogger.Println("count metrics with id statement prepare error:", err)
		}
		insertCounterStmt, err = db.Prepare(`
			INSERT INTO metrics (id, type, value, delta) VALUES ($1, 'counter', NULL, $2)
		`)
		if err != nil {
			loggers.ErrorLogger.Println("insert counter statement prepare error:", err)
		}
		updateCounterStmt, err = db.Prepare(`
			UPDATE metrics SET delta=$2 WHERE id=$1;
		`)
		if err != nil {
			loggers.ErrorLogger.Println("update counter statement prepare error:", err)
		}
		insertGaugeStmt, err = db.Prepare(`
			INSERT INTO metrics (id, type, value, delta) VALUES ($1, 'gauge', $2, NULL)
			ON CONFLICT (id, type) DO UPDATE SET
				value=$2,
				delta=NULL;
		`)
		if err != nil {
			loggers.ErrorLogger.Println("insert statement prepare error:", err)
		}
		selectAllStmt, err = db.Prepare(`SELECT id, type, value, delta FROM metrics;`)
		if err != nil {
			loggers.ErrorLogger.Println("select all statement prepare error:", err)
		}
		selectOneGaugeStmt, err = db.Prepare(`SELECT value FROM metrics WHERE id=$1;`)
		if err != nil {
			loggers.ErrorLogger.Println("select one gauge statement prepare error:", err)
		}
		selectOneCounterStmt, err = db.Prepare(`SELECT delta FROM metrics WHERE id=$1;`)
		if err != nil {
			loggers.ErrorLogger.Println("select one counter statement prepare error:", err)
		}
	}
	return Database{
		DB:                               db,
		InsertCounterToDatabaseStmt:      insertCounterStmt,
		UpdateCounterToDatabaseStmt:      updateCounterStmt,
		InsertUpdateGaugeToDatabaseStmt:  insertGaugeStmt,
		SelectAllFromDatabaseStmt:        selectAllStmt,
		SelectOneGaugeFromDatabaseStmt:   selectOneGaugeStmt,
		SelectOneCounterFromDatabaseStmt: selectOneCounterStmt,
		CountIDsInDatabaseStmt:           countIDsStmt,
	}
}

func SetDatabase(db *sql.DB, dbAddress string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/server/database/migrations",
		dbAddress, driver)
	if err != nil {
		return fmt.Errorf("could not create migration: %w", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db Database) GetAllMetrics() ([]types.Metrics, error) {
	var metrics []types.Metrics
	rows, err := db.DB.Query("SELECT metrics.id, type, value, delta FROM metrics")
	if err != nil {
		return nil, fmt.Errorf("error while getting metric from database: %w", err)
	}
	for rows.Next() {
		var m types.Metrics
		rows.Scan(&m.ID, &m.MType, m.Delta, m.Value)
		metrics = append(metrics, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error while scanning metrics from database: %w", err)
	}
	return metrics, nil
}

func (db Database) GetMetric(m types.Metrics, key string) (types.Metrics, error) {
	switch m.MType {
	case "gauge":
		var value float64
		err := db.SelectOneGaugeFromDatabaseStmt.QueryRow(m.ID).Scan(&value)
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
		var delta int64
		err := db.SelectOneCounterFromDatabaseStmt.QueryRow(m.ID).Scan(&delta)
		if err != nil {
			return m, myerrors.ErrTypeNotFound
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
		_, err := db.InsertUpdateGaugeToDatabaseStmt.Exec(m.ID, *m.Value)
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
		var numberOfMetrics int
		err := db.CountIDsInDatabaseStmt.QueryRow(m.ID).Scan(&numberOfMetrics)
		if err != nil {
			return err
		}
		if numberOfMetrics != 0 {
			var delta int64
			err = db.SelectOneCounterFromDatabaseStmt.QueryRow(m.ID).Scan(&delta)
			if err != nil {
				return err
			}
			_, err = db.UpdateCounterToDatabaseStmt.Exec(m.ID, delta+*m.Delta)
			if err != nil {
				return err
			}
		} else {
			_, err = db.InsertCounterToDatabaseStmt.Exec(m.ID, *m.Delta)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%wno such type of metric", myerrors.ErrTypeNotImplemented)
	}
	return nil
}

func (db Database) Check() error {
	return db.DB.Ping()
}

func (db Database) SaveManyMetrics(metrics []types.Metrics, key string) error {
	for _, m := range metrics {
		err := db.SaveMetric(m, key)
		if err != nil {
			return err
		}
	}
	return nil
}
