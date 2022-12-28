package agent

import "time"

const (
	Protocol    = "http://"
	Server      = "127.0.0.1"
	Port        = "8080"
	ContentType = "text/plain"
	TCP         = "tcp"
	Timeout     = 2 * time.Second
)

var Metrics []Gauge
var PollCount int64
