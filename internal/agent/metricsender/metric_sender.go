// // Package metricsender describes sending metrics' info to the server
package metricsender

import "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"

// MetricSender sends metrics to the server
type MetricSender interface {

	//SendAllMetrics sends metrics to the server one by one
	SendAllMetrics(c *metriccollector.MetricCollector)

	//SendAllMetricsAsButch sends metrics to the server at once
	SendAllMetricsAsButch(c *metriccollector.MetricCollector)
}
