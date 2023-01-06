package agent

import "time"

const (
	DefaultProtocol      = "http://"
	DefaultHost          = "127.0.0.1"
	DefaultPort          = "8080"
	ContentTypeTextPlain = "text/plain"
	TCP                  = "tcp"
	DefaultTimeout       = 2 * time.Second
)

var allMetrics []Gauge
var PollCount int64
