package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	HTTPResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "HTTP response size in bytes",
		Buckets: prometheus.ExponentialBuckets(100, 10, 5),
	}, []string{"method", "path"})

	// gRPC метрики
	GRPCRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	}, []string{"method", "status"})

	GRPCRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "gRPC request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "status"})

	// DB метрики
	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	DBActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_active_connections",
		Help: "Number of active database connections",
	})

	DBIdleConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "Number of idle database connections",
	})

	// метрики для агрегатора
	AggregatorPacketsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "aggregator_packets_received_total",
		Help: "Total number of packets received by the aggregator",
	})

	AggregatorPacketsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "aggregator_packets_processed_total",
		Help: "Total number of packets successfully processed",
	})

	AggregatorPacketsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "aggregator_packets_failed_total",
		Help: "Total number of packets failed during processing",
	})

	AggregatorPacketProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "aggregator_packet_processing_seconds",
		Help:    "Histogram of packet processing durations",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // от 1ms до ~16 секунд
	})

	AggregatorActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aggregator_active_workers",
		Help: "Current number of active workers processing packets",
	})
)
