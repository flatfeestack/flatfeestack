package prom

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
)

var Registry *prometheus.Registry

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	Registry = prometheus.NewRegistry()
	Registry.MustRegister(collectors.NewBuildInfoCollector())
	Registry.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
	))
	Registry.MustRegister(TotalRequests)
	Registry.MustRegister(ResponseStatus)
	Registry.MustRegister(HttpDuration)
}

var TotalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var ResponseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(HttpDuration.WithLabelValues(path))
		rw := &responseWriterWrapper{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		ResponseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		TotalRequests.WithLabelValues(path).Inc()

		timer.ObserveDuration()
	})
}
