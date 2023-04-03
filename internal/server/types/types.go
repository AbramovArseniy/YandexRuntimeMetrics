package types

import (
	"compress/gzip"
	"net/http"
)

// Metrics stores metric data
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
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
	StorageTypeDB   StorageType = "database"
	StorageTypeFile StorageType = "file"
)
