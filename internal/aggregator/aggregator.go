package aggregator

import (
	"context"
	"sync"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"
	"github.com/CoolE88/data-aggregation-service/internal/metrics"

	"go.uber.org/zap"
)

type DataService interface {
	ProcessPacket(ctx context.Context, packet *domain.DataPacket) error
}

type Aggregator struct {
	workers int
	service DataService
	logger  *zap.Logger
}

func NewAggregator(service DataService, workers int, logger *zap.Logger) *Aggregator {
	return &Aggregator{
		service: service,
		workers: workers,
		logger:  logger,
	}
}

func (a *Aggregator) Start(ctx context.Context, packets chan *domain.DataPacket) {
	a.logger.Info("starting aggregator",
		zap.Int("workers", a.workers),
		zap.String("start_time", time.Now().Format(time.RFC3339)),
	)

	// Устанавливаем начальное количество активных воркеров
	metrics.AggregatorActiveWorkers.Set(float64(a.workers))

	var wg sync.WaitGroup

	for i := 0; i < a.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			// При завершении воркера уменьшаем счетчик
			defer metrics.AggregatorActiveWorkers.Dec()

			a.logger.Debug("worker started",
				zap.Int("worker_id", workerID),
			)

			for {
				select {
				case packet, ok := <-packets:
					if !ok {
						a.logger.Info("packet channel closed, exiting worker",
							zap.Int("worker_id", workerID),
						)
						return
					}

					select {
					case <-ctx.Done():
						a.logger.Info("context cancelled, exiting worker",
							zap.Int("worker_id", workerID),
						)
						return
					default:
						// Продолжаем обработку, если контекст не отменен
					}

					// Увеличиваем счётчик полученных пакетов
					metrics.AggregatorPacketsReceived.Inc()
					startTime := time.Now()

					err := a.service.ProcessPacket(ctx, packet)
					if err != nil {
						// Увеличиваем счётчик неудачных пакетов
						metrics.AggregatorPacketsFailed.Inc()

						a.logger.Error("[Aggregator] failed to process packet",
							zap.Int("worker_id", workerID),
							zap.String("packet_id", packet.ID.String()),
							zap.String("packet_timestamp", packet.Timestamp.Format(time.RFC3339Nano)),
							zap.Ints("payload", packet.Payload),
							zap.Error(err),
						)
						continue
					}

					// Увеличиваем счётчик успешно обработанных пакетов
					metrics.AggregatorPacketsProcessed.Inc()

					processingTime := time.Since(startTime)
					// Записываем время обработки в гистограмму
					metrics.AggregatorPacketProcessingTime.Observe(processingTime.Seconds())

					a.logger.Info("[Aggregator] packet processed successfully",
						zap.Int("worker_id", workerID),
						zap.String("packet_id", packet.ID.String()),
						zap.String("packet_timestamp", packet.Timestamp.Format(time.RFC3339Nano)),
						zap.Ints("payload", packet.Payload),
						zap.Duration("processing_time", processingTime),
					)

				case <-ctx.Done():
					a.logger.Info("context cancelled, exiting worker",
						zap.Int("worker_id", workerID),
					)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// После завершения всех воркеров обнуляем счетчик
	metrics.AggregatorActiveWorkers.Set(0)

	a.logger.Info("aggregator stopped",
		zap.String("stop_time", time.Now().Format(time.RFC3339)),
	)
}

// Stop аккуратная остановка агрегатора
func (a *Aggregator) Stop() {
	a.logger.Info("stopping aggregator gracefully")
}
