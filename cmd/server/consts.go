package main

import "time"

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultRestore       = true
	createTableQuerySQL  = `
				CREATE TABLE IF NOT EXISTS metrics (
					id VARCHAR(128) PRIMARY KEY,
					type VARCHAR(32) NOT NULL,
					value DOUBLE PRECISION,
					delta BIGINT
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_metrics_id_type ON metrics (id, type);
		`
)
