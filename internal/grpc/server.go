package grpc

import (
	"context"
	"net"
	"time"

	pb "github.com/CoolE88/data-aggregation-service/gen/go/aggregator/v1"
	"github.com/CoolE88/data-aggregation-service/internal/domain"
	"github.com/CoolE88/data-aggregation-service/internal/metrics"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// DataService описывает бизнес-логику для получения данных
type DataService interface {
	GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error)
	GetMaxValueByPacketID(ctx context.Context, packetID string) (*domain.ProcessedData, error)
	CheckDBConnection(ctx context.Context) error
}

// GRPCServer реализует gRPC сервер с метриками и логированием
type GRPCServer struct {
	pb.UnimplementedDataAggregationServiceServer
	server  *grpc.Server
	service DataService
	logger  *zap.Logger
}

func NewGRPCServer(service DataService, logger *zap.Logger) *GRPCServer {
	loggingInterceptor := logging.UnaryServerInterceptor(interceptorLogger(logger))
	metricsInterceptor := grpc_prometheus.UnaryServerInterceptor
	customMetricsInterceptor := unaryMetricsInterceptor()

	chain := grpc.ChainUnaryInterceptor(
		loggingInterceptor,
		metricsInterceptor,
		customMetricsInterceptor,
	)

	s := &GRPCServer{
		server:  grpc.NewServer(chain),
		service: service,
		logger:  logger,
	}

	pb.RegisterDataAggregationServiceServer(s.server, s)
	reflection.Register(s.server)

	grpc_prometheus.Register(s.server)
	grpc_prometheus.EnableHandlingTimeHistogram()

	return s
}

func (s *GRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("Starting gRPC server", zap.String("addr", addr))
	return s.server.Serve(lis)
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down gRPC server")

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}

// Custom metrics interceptor для детального отслеживания статусов и длительности с статусом
func unaryMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		var statusCode string
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = codes.Unknown.String()
			}
		} else {
			statusCode = codes.OK.String()
		}

		duration := time.Since(start).Seconds()

		metrics.GRPCRequests.WithLabelValues(info.FullMethod, statusCode).Inc()
		metrics.GRPCRequestDuration.WithLabelValues(info.FullMethod, statusCode).Observe(duration)

		return resp, err
	}
}

// Logger adapter для grpc middleware
func interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)
		for i := 0; i < len(fields); i += 2 {
			key := fields[i].(string)
			value := fields[i+1]
			f = append(f, zap.Any(key, value))
		}
		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			logger.Info(msg)
		}
	})
}

func (s *GRPCServer) GetMaxValuesByPeriod(ctx context.Context, req *pb.TimePeriod) (*pb.MaxValuesResponse, error) {
	if req.StartTime == "" || req.EndTime == "" {
		return nil, status.Error(codes.InvalidArgument, "start_time and end_time are required")
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start_time format, expected RFC3339")
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid end_time format, expected RFC3339")
	}

	data, err := s.service.GetMaxValuesByTimeRange(ctx, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get max values by period", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to retrieve data")
	}

	response := &pb.MaxValuesResponse{
		MaxValues: make([]*pb.MaxValue, len(data)),
	}

	for i, item := range data {
		response.MaxValues[i] = &pb.MaxValue{
			Id:       item.PacketID.String(),
			MaxValue: int32(item.MaxValue),
		}
	}

	return response, nil
}

func (s *GRPCServer) GetMaxValueByID(ctx context.Context, req *pb.PackageID) (*pb.MaxValueResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	data, err := s.service.GetMaxValueByPacketID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get max value by ID", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to retrieve packet")
	}

	if data == nil {
		return nil, status.Error(codes.NotFound, "packet not found")
	}

	response := &pb.MaxValueResponse{
		Id:       data.PacketID.String(),
		MaxValue: int32(data.MaxValue),
	}

	return response, nil
}
