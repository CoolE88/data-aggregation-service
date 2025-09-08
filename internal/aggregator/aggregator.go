package aggregator

import (
	"context"
	"sync"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"
	"github.com/CoolE88/data-aggregation-service/internal/metrics"
	"github.com/CoolE88/data-aggregation-service/pkg/utils"

	"go.uber.org/zap"
)

type DataService interface {
	ProcessPacket(ctx context.Context, packet *domain.DataPacket) error
}

type Aggregator struct {
	workers int
	service DataService
	logger  *zap.Logger
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewAggregator(service DataService, workers int, logger *zap.Logger) *Aggregator {
	return &Aggregator{
		service: service,
		workers: workers,
		logger:  logger,
	}
}

func (a *Aggregator) Start(ctx context.Context, packets chan *domain.DataPacket) {
	aggCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	a.wg.Add(a.workers)

	for i := 0; i < a.workers; i++ {
		go a.worker(aggCtx, packets, &a.wg, i)
	}

	go func() {
		a.wg.Wait()
		a.logger.Info("All aggregator workers finished")
	}()
}

func (a *Aggregator) Wait() {
	a.wg.Wait() // Для внешнего ожидания
}

func (a *Aggregator) worker(ctx context.Context, packets chan *domain.DataPacket, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	a.logger.Info("Worker started", zap.Int("worker_id", id))
	metrics.AggregatorActiveWorkers.Inc()
	defer func() {
		metrics.AggregatorActiveWorkers.Dec()
		a.logger.Info("Worker stopped", zap.Int("worker_id", id))
	}()

	for {
		select {
		case packet, ok := <-packets:
			if !ok {
				a.logger.Info("Packets channel closed, stopping worker", zap.Int("worker_id", id))
				return
			}
			metrics.AggregatorPacketsReceived.Inc()

			// Валидация UUID
			if err := utils.IsValidUUID(packet.ID.String()); err != nil {
				metrics.AggregatorPacketsFailed.Inc()
				a.logger.Error("Invalid UUID in packet", zap.String("packet_id", packet.ID.String()), zap.Error(err), zap.Int("worker_id", id))
				continue
			}

			if ctx.Err() != nil { // Проверка контекста перед обработкой
				a.logger.Info("Context cancelled before processing packet", zap.Int("worker_id", id))
				return
			}

			start := time.Now()
			err := a.service.ProcessPacket(ctx, packet)
			duration := time.Since(start).Seconds()
			metrics.AggregatorPacketProcessingTime.Observe(duration)

			if err != nil {
				metrics.AggregatorPacketsFailed.Inc()
				a.logger.Error("Failed to process packet", zap.Error(err), zap.Int("worker_id", id))
			} else {
				metrics.AggregatorPacketsProcessed.Inc()
				a.logger.Debug("Packet processed", zap.Duration("duration", time.Duration(duration*float64(time.Second))), zap.Int("worker_id", id))
			}
		case <-ctx.Done():
			a.logger.Info("Context cancelled, stopping worker", zap.Int("worker_id", id))
			return
		}
	}
}

func (a *Aggregator) Stop() {
	if a.cancel != nil {
		a.cancel()
		a.logger.Info("Aggregator context cancelled via Stop()")
		a.wg.Wait() // Дождаться завершения воркеров
	}
}
