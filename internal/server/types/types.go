package types

import (
	"compress/gzip"
	"net/http"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type GZIPWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w GZIPWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type StorageType string

const (
	StorageTypeDB   StorageType = "database"
	StorageTypeFile StorageType = "file"
)
