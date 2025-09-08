package service

import (
	"context"
	"fmt"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	SaveProcessedData(ctx context.Context, data *domain.ProcessedData) error
	GetMaxValueByPacketID(ctx context.Context, packetID uuid.UUID) (*domain.ProcessedData, error)
	GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error)
	HealthCheck(ctx context.Context) error
}

type DataService struct {
	repo   Repository
	logger *zap.Logger
}

func (s *DataService) CheckDBConnection(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

func NewDataService(repo Repository, logger *zap.Logger) *DataService {
	return &DataService{
		repo:   repo,
		logger: logger,
	}
}

// ProcessPacket находит максимальное число из пакетного пейлода
func (s *DataService) ProcessPacket(ctx context.Context, packet *domain.DataPacket) error {
	if err := ctx.Err(); err != nil {
		s.logger.Warn("[DataService] Processing cancelled by context",
			zap.String("packet_id", packet.ID.String()))
		return ctx.Err()
	}

	maxValue := s.FindMaxValue(packet.Payload)

	processedData := &domain.ProcessedData{
		PacketID:        packet.ID,
		PacketCreatedAt: packet.Timestamp, // timestamp из пакета
		CreatedAt:       time.Now().UTC(), // время обработки в UTC
		MaxValue:        maxValue,
	}

	if err := s.repo.SaveProcessedData(ctx, processedData); err != nil {
		s.logger.Error("[DataService] Failed to save processed data",
			zap.String("packet_id", packet.ID.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("[DataService] Packet processed successfully",
		zap.String("packet_id", packet.ID.String()),
		zap.Int("max_value", maxValue))

	return nil
}

func (s *DataService) FindMaxValue(payload []int) int {
	if len(payload) == 0 {
		return 0
	}

	max := payload[0]
	for _, value := range payload {
		if value > max {
			max = value
		}
	}
	return max
}

// GetMaxValueByPacketID возвращает запись с максимальным значением по заданному packetID.
// Если запись не найдена, возвращает (nil, nil).
func (s *DataService) GetMaxValueByPacketID(ctx context.Context, packetID string) (*domain.ProcessedData, error) {
	id, err := uuid.Parse(packetID)
	if err != nil {
		return nil, fmt.Errorf("invalid packet ID: %w", err) // оборачиваем ошибку
	}

	data, err := s.repo.GetMaxValueByPacketID(ctx, id)
	if err != nil {
		s.logger.Error("[DataService] Failed to get max value by packet ID",
			zap.String("packet_id", packetID),
			zap.Error(err))
		return nil, err
	}

	return data, nil
}

// GetMaxValuesByTimeRange возвращает запись с максимальным значением по заданному временному интервалу
func (s *DataService) GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	data, err := s.repo.GetMaxValuesByTimeRange(ctx, start, end)
	if err != nil {
		s.logger.Error("[DataService] Failed to get max values by time range",
			zap.Time("start", start),
			zap.Time("end", end),
			zap.Error(err))
		return nil, err
	}

	return data, nil
}
