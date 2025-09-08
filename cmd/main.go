package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/aggregator"
	"github.com/CoolE88/data-aggregation-service/internal/config"
	"github.com/CoolE88/data-aggregation-service/internal/domain"
	appgrpc "github.com/CoolE88/data-aggregation-service/internal/grpc"
	apphttp "github.com/CoolE88/data-aggregation-service/internal/http"
	applogger "github.com/CoolE88/data-aggregation-service/internal/logger"
	"github.com/CoolE88/data-aggregation-service/internal/repository/postgres"
	"github.com/CoolE88/data-aggregation-service/internal/service"
	"github.com/CoolE88/data-aggregation-service/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	// Создаём отменяемый контекст для всего приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Гарантирует отмену при выходе

	cfg := config.LoadConfig()

	logger, err := applogger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error during logger sync: %v", err)
		}
	}()

	logger.Info("Starting Data Aggregation Service", zap.String("version", "1.0.0"))

	// Инициализация репозитория
	repo, err := postgres.NewPostgresRepository(ctx, cfg.DBConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return
	}
	defer func() {
		repo.Close()
		logger.Info("Database connection closed")
	}()

	logger.Info("Database connection established")

	// Инициализация сервиса
	dataService := service.NewDataService(repo, logger)

	// Запуск HTTP сервера
	httpServer := apphttp.NewHTTPServer(cfg.RESTPort, dataService, logger)
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
			return
		}
	}()

	// Запуск GRPC сервера
	grpcServer := appgrpc.NewGRPCServer(dataService, logger)
	go func() {
		if err := grpcServer.Start(cfg.GRPCPort); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
			return
		}
	}()

	// Инициализация агрегатора
	packets := make(chan *domain.DataPacket, 1000)
	aggregator := aggregator.NewAggregator(dataService, cfg.WorkerCount, logger)

	// Запускаем агрегатор
	go func() {
		aggregator.Start(ctx, packets)
	}()

	// Генерация пакетов
	timeGenerator := utils.PartitionedTimeGenerator()
	ticker := time.NewTicker(time.Duration(cfg.DataInterval) * time.Millisecond)
	defer ticker.Stop()

	logger.Info("Starting packet generation")

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)

	// Имитируем внешний источник с пакетами
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var closed bool

		for {
			select {
			case <-ticker.C:
				if ctx.Err() != nil { // Проверка контекста перед генерацией
					logger.Info("Context cancelled, stopping packet generation")
					if !closed {
						close(packets)
						closed = true //nolint:ineffassign // to avoid double closing
					}
					return
				}
				packet := &domain.DataPacket{
					ID:        utils.NewUUID(),
					Timestamp: timeGenerator.Generate(),
					Payload:   utils.GenerateRandomPayload(10),
				}

				select {
				case packets <- packet:
					logger.Debug("Generated new packet", zap.String("packet_id", packet.ID.String()))
				default:
					logger.Warn("Packet channel full, dropping packet")
				}

			case <-ctx.Done():
				logger.Info("Stopping packet generation due to context cancellation")
				if !closed {
					close(packets)
					closed = true //nolint:ineffassign // to avoid double closing
				}
				return
			}
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down servers...")

	// Отменяем контекст для всех компонентов (остановит генерацию, мониторинг и агрегатор)
	cancel()

	// Останавливаем генерацию
	ticker.Stop()

	// Останавливаем агрегатор явно
	aggregator.Stop()
	aggregator.Wait() // Дождаться завершения воркеров
	wg.Wait()         // Дождаться генератора

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	// Останавливаем GRPC сервер
	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("gRPC server shutdown due to timeout")
		} else {
			logger.Error("gRPC server shutdown failed", zap.Error(err))
		}
	}

	logger.Info("Data Aggregation Service stopped")
}
