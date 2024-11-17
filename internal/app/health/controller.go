package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type serviceMiddleware struct {
	fn   log.Logger
	next Service
}

// It represents a single RPC. That is, a single method in our service interface.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)
type Middleware func(next Endpoint) Endpoint

type Controller struct {
	prom             *prometheus.Registry // middleware!!
	tracer           trace.Tracer
	readinessHandler func() bool // service from registry
}

func New(prom *prometheus.Registry, readinessHandler func() bool) *Controller {
	return &Controller{
		prom:             prom,
		readinessHandler: readinessHandler,
		tracer:           otel.GetTracerProvider().Tracer("health_controller"),
	}
}

func (c *Controller) RegisterRoutes(router *echo.Group) {
	router.GET("/health", c.Health)
	router.GET("/readiness", c.Readiness)
	router.GET("/metrics", c.PrometheusHandler())
}

// Health - endpoint for health
// //nolint:revive
func (c *Controller) Health(ectx echo.Context) error {
	return ectx.JSON(http.StatusOK, echo.Map{"status": "healthy"})
}
func (c *Controller) Readiness(ectx echo.Context) error {
	// decode request -> endpoint(call service) -> encode response
	if c.readinessHandler() {
		return ectx.JSON(http.StatusOK, echo.Map{"status": "ready"})
	}
	return ectx.JSON(http.StatusServiceUnavailable, echo.Map{"status": "not ready"})
}
func (c *Controller) PrometheusHandler() echo.HandlerFunc {
	if c.prom != nil {
		h := promhttp.HandlerFor(c.prom, promhttp.HandlerOpts{
			Registry:          c.prom,
			EnableOpenMetrics: true,
		})
		return func(ectx echo.Context) error {
			h.ServeHTTP(ectx.Response(), ectx.Request())
			return nil
		}
	}
	return func(ectx echo.Context) error {
		return nil
	}
}
