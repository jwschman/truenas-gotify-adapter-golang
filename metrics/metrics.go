package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	startTime = time.Now()

	// compute uptime on demand
	Uptime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Time in seconds since the application started.",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)

	RequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "Total number of requests received",
		},
	)

	RequestsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_requests_failed_total",
			Help: "Total number of failed requests",
		},
	)

	GotifySendsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_gotify_sends_total",
			Help: "Total number of notifications sent to Gotify",
		},
	)

	GotifySendsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_gotify_sends_failed_total",
			Help: "Total number of notifications failed to send to Gotify",
		},
	)

	RequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_requests_duration_seconds",
			Help:    "Duration of handling incoming requests in seconds",
			Buckets: prometheus.DefBuckets, // just use default buckets
		},
	)

	GotifySendDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_gotify_send_duration_seconds",
			Help:    "Duration of Gotify notification send operations in seconds.",
			Buckets: prometheus.DefBuckets, // default buckets here as well
		},
	)
)

// Register all defined metrics
func Register() {
	prometheus.MustRegister(Uptime)
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestsFailedTotal)
	prometheus.MustRegister(GotifySendsTotal)
	prometheus.MustRegister(GotifySendsFailedTotal)
	prometheus.MustRegister(GotifySendDuration)
	prometheus.MustRegister(RequestDuration)
}
