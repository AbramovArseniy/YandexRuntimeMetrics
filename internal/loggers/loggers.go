// Package loggers logs events
package loggers

import (
	"log"
	"os"
)

// Types of loggers
var (
	DebugLogger = log.New(os.Stdout, "DEBUG \t", log.LstdFlags)
	ErrorLogger = log.New(os.Stdout, "ERROR \t", log.LstdFlags)
	InfoLogger  = log.New(os.Stdout, "INFO \t", log.LstdFlags)
)
