// Package types contains variable types
package types

import (
	"compress/gzip"
	"net/http"
)

// Metrics stores metric data
type Metrics struct {
	ID    string   `json:"id" redis:"id"`
	MType string   `json:"type" redis:"type"`
	Delta *int64   `json:"delta,omitempty" redis:"delta"`
	Value *float64 `json:"value,omitempty" redis:"value"`
	Hash  string   `json:"hash,omitempty" redis:"hash"`
}

// GZIPWriter writes http response encoded as gzip
type GZIPWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

// Write writes data encoded as gzip
func (w GZIPWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// StorageType stores type of storage
type StorageType string

// Types of storage
const (
	StorageTypePostgres StorageType = "postgres"
	StorageTypeFile     StorageType = "file"
	StorageTypeRedis    StorageType = "redis"
)
