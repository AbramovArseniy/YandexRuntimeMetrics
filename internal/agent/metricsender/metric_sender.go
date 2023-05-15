package metricsender

import "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/agent/metriccollector"

type MetricSender interface {
	SendAllMetrics(c *metriccollector.MetricCollector)
	SendAllMetricsAsButch(c *metriccollector.MetricCollector)
}
