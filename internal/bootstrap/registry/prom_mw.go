package registry

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMiddleware struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewPrometheusMiddleware(reg *prometheus.Registry) *PrometheusMiddleware {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by status code and path",
		},
		[]string{"status", "path", "method"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"status", "path", "method"},
	)

	reg.MustRegister(requestCounter)
	reg.MustRegister(requestDuration)

	return &PrometheusMiddleware{
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

func (p *PrometheusMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			status := strconv.Itoa(c.Response().Status)

			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}

			method := c.Request().Method

			p.requestCounter.WithLabelValues(status, path, method).Inc()

			duration := time.Since(start).Seconds()
			p.requestDuration.WithLabelValues(status, path, method).Observe(duration)

			return err
		}
	}
}
