package agent

const (
	DefaultProtocol = "http://"
	DefaultHost     = "127.0.0.1"
	DefaultPort     = "8080"
)

var allMetrics []Gauge
var PollCount int64
