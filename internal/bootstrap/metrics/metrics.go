package metrics

import "github.com/prometheus/client_golang/prometheus"

type MetricsCollector struct {
	healthCheckCounter *prometheus.CounterVec
	registry           *prometheus.Registry
}

func NewMetricsCollector(reg *prometheus.Registry) *MetricsCollector {
	m := &MetricsCollector{
		healthCheckCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "health_check_status_total",
				Help: "Total number of health checks by status",
			},
			[]string{"checker", "status"},
		),
		registry: reg,
	}

	reg.MustRegister(m.healthCheckCounter)
	return m
}

func (m *MetricsCollector) RecordHealthCheck(name string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	m.healthCheckCounter.WithLabelValues(name, status).Inc()
}
