package metrics

import (
	"net/http"
	"strconv"

	"github.com/cabewaldrop/liminal/pkg/middleware/capture"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type metrics struct {
	totalRequestCount prometheus.Counter
	requestCount      *prometheus.CounterVec
	responseStatus    *prometheus.CounterVec
	next              http.Handler
}

func NewMetrics(next http.Handler, reg prometheus.Registerer) *metrics {
	totalRequestCount := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_request_count",
			Help: "The total number of requests across all paths",
		})

	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_count",
		Help: "The request count partitioned by path",
	}, []string{"path"})

	responseStatus := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "response_status",
		Help: "The count of HTTP response codes partitioned by path",
	}, []string{"path", "status_code"})

	reg.MustRegister(totalRequestCount, requestCount, responseStatus)
	return &metrics{
		totalRequestCount: totalRequestCount,
		requestCount:      requestCount,
		responseStatus:    responseStatus,
		next:              next,
	}
}

func (m *metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	labels := make(prometheus.Labels)
	labels["path"] = r.URL.Path
	m.totalRequestCount.Inc()
	m.requestCount.With(labels).Inc()

	capture := capture.NewResponseWriter(w)
	m.next.ServeHTTP(capture, r)

	log.Logger.Info().Msgf("Got back the following status: %d", capture.Status())
	labels["status_code"] = strconv.Itoa(capture.Status())
	m.responseStatus.With(labels).Inc()
}
