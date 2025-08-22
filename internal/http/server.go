package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"
	"github.com/CoolE88/data-aggregation-service/internal/metrics"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type DataService interface {
	GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error)
	GetMaxValueByPacketID(ctx context.Context, packetID string) (*domain.ProcessedData, error)
	CheckDBConnection(ctx context.Context) error
}

type HTTPServer struct {
	server  *http.Server
	service DataService
	logger  *zap.Logger
}

func NewHTTPServer(addr string, service DataService, logger *zap.Logger) *HTTPServer {
	router := mux.NewRouter()

	s := &HTTPServer{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		service: service,
		logger:  logger,
	}

	// Middleware регистрации
	router.Use(s.metricsMiddleware)
	router.Use(s.loggingMiddleware)

	// Маршруты
	router.HandleFunc("/health", s.healthCheck).Methods("GET")
	router.HandleFunc("/api/v1/max-values", s.getMaxValuesByTimeRange).Methods("GET")
	router.HandleFunc("/api/v1/max-values/{id}", s.getMaxValueByID).Methods("GET")

	// Метрики Prometheus
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	return s
}

func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// responseWriter для отслеживания статус кода и размера
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// middleware для сбора метрик HTTP запросов с использованием шаблона пути
func (s *HTTPServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		method := r.Method
		status := strconv.Itoa(rw.statusCode)

		// Получаем шаблон пути из mux (если доступен)
		path := r.URL.Path
		if route := mux.CurrentRoute(r); route != nil {
			if tpl, err := route.GetPathTemplate(); err == nil {
				path = tpl
			}
		}

		metrics.HTTPRequests.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(rw.size))
	})
}

// middleware для логирования HTTP запросов
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		s.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("status", rw.statusCode),
			zap.Int("response_size", rw.size),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func (s *HTTPServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.service.CheckDBConnection(r.Context()); err != nil {
		s.logger.Error("Health check failed", zap.Error(err))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		s.logger.Error("Failed to encode health check response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getMaxValuesByTimeRange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		s.logger.Error("invalid start time format",
			zap.Error(err),
			zap.String("received_start", startStr))
		http.Error(w, "invalid start time format", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		s.logger.Error("invalid end time format",
			zap.Error(err),
			zap.String("received_end", endStr))
		http.Error(w, "invalid end time format", http.StatusBadRequest)
		return
	}

	data, err := s.service.GetMaxValuesByTimeRange(ctx, start, end)
	if err != nil {
		s.logger.Error("Failed to get max values by time range", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getMaxValueByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	data, err := s.service.GetMaxValueByPacketID(r.Context(), id)
	if err != nil {
		s.logger.Error("Failed to get max value by ID", zap.String("id", id), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if data == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
