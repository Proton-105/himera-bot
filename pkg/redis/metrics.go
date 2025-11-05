package redis

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	goredis "github.com/redis/go-redis/v9"
)

var (
	redisRequestsTotal   *prometheus.CounterVec
	redisErrorsTotal     *prometheus.CounterVec
	redisRequestDuration *prometheus.HistogramVec
)

func init() {
	redisRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_requests_total",
			Help: "Total number of Redis requests by method.",
		},
		[]string{"method"},
	)
	redisErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_errors_total",
			Help: "Total number of Redis errors by method.",
		},
		[]string{"method"},
	)
	redisRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_request_duration_seconds",
			Help:    "Redis request latency distributions.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(redisRequestsTotal, redisErrorsTotal, redisRequestDuration)
}

// MetricsClient wraps Client to collect Prometheus metrics.
type MetricsClient struct {
	next *Client
}

// NewMetricsClient creates an instrumented Redis client.
func NewMetricsClient(next *Client) *MetricsClient {
	return &MetricsClient{next: next}
}

// Get instruments Client.Get.
func (m *MetricsClient) Get(ctx context.Context, key string) (string, error) {
	timer := prometheus.NewTimer(redisRequestDuration.WithLabelValues("get"))
	result, err := m.next.Get(ctx, key)
	timer.ObserveDuration()
	redisRequestsTotal.WithLabelValues("get").Inc()
	if err != nil {
		redisErrorsTotal.WithLabelValues("get").Inc()
	}
	return result, err
}

// Set instruments Client.Set.
func (m *MetricsClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	timer := prometheus.NewTimer(redisRequestDuration.WithLabelValues("set"))
	err := m.next.Set(ctx, key, value, ttl)
	timer.ObserveDuration()
	redisRequestsTotal.WithLabelValues("set").Inc()
	if err != nil {
		redisErrorsTotal.WithLabelValues("set").Inc()
	}
	return err
}

// Delete instruments Client.Delete.
func (m *MetricsClient) Delete(ctx context.Context, key string) error {
	timer := prometheus.NewTimer(redisRequestDuration.WithLabelValues("delete"))
	err := m.next.Delete(ctx, key)
	timer.ObserveDuration()
	redisRequestsTotal.WithLabelValues("delete").Inc()
	if err != nil {
		redisErrorsTotal.WithLabelValues("delete").Inc()
	}
	return err
}

// Close closes underlying client.
func (m *MetricsClient) Close() error {
	return m.next.Close()
}

// TxPipeline forwards to the underlying client.
func (m *MetricsClient) TxPipeline() goredis.Pipeliner {
	return m.next.TxPipeline()
}
