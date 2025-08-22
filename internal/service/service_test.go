package service

import (
	"context"
	"testing"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveProcessedData(ctx context.Context, data *domain.ProcessedData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRepository) GetMaxValueByPacketID(ctx context.Context, packetID uuid.UUID) (*domain.ProcessedData, error) {
	args := m.Called(ctx, packetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProcessedData), args.Error(1)
}

func (m *MockRepository) GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error) {
	args := m.Called(ctx, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ProcessedData), args.Error(1)
}

func (m *MockRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestDataService_FindMaxValue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &DataService{logger: logger}

	tests := []struct {
		name     string
		payload  []int
		expected int
	}{
		{"positive numbers", []int{1, 5, 3, 9, 2}, 9},
		{"negative numbers", []int{-1, -5, -3}, -1},
		{"mixed numbers", []int{-1, 5, 0, -10}, 5},
		{"single element", []int{42}, 42},
		{"empty slice", []int{}, 0},
		{"zeros", []int{0, 0, 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FindMaxValue(tt.payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataService_ProcessPacket_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	service := NewDataService(mockRepo, logger)

	packet := &domain.DataPacket{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Payload:   []int{1, 5, 3, 9, 2},
	}

	expectedData := &domain.ProcessedData{
		PacketID:        packet.ID,
		PacketCreatedAt: packet.Timestamp,
		MaxValue:        9,
		CreatedAt:       time.Now(),
	}

	mockRepo.On("SaveProcessedData", mock.Anything, mock.AnythingOfType("*domain.ProcessedData")).
		Return(nil).
		Run(func(args mock.Arguments) {
			data := args.Get(1).(*domain.ProcessedData)
			assert.Equal(t, expectedData.PacketID, data.PacketID)
			assert.Equal(t, expectedData.MaxValue, data.MaxValue)
		})

	err := service.ProcessPacket(context.Background(), packet)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDataService_GetMaxValueByPacketID_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	service := NewDataService(mockRepo, logger)

	packetID := uuid.New()
	expectedData := &domain.ProcessedData{
		PacketID:        packetID,
		PacketCreatedAt: time.Now(),
		MaxValue:        42,
		CreatedAt:       time.Now(),
	}

	mockRepo.On("GetMaxValueByPacketID", mock.Anything, packetID).
		Return(expectedData, nil)

	result, err := service.GetMaxValueByPacketID(context.Background(), packetID.String())
	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	mockRepo.AssertExpectations(t)
}

func TestDataService_GetMaxValueByPacketID_InvalidID(t *testing.T) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	service := NewDataService(mockRepo, logger)

	result, err := service.GetMaxValueByPacketID(context.Background(), "invalid-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid packet ID")
}

func TestDataService_GetMaxValuesByTimeRange_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	service := NewDataService(mockRepo, logger)

	start := time.Now().Add(-time.Hour)
	end := time.Now()
	expectedData := []*domain.ProcessedData{
		{PacketID: uuid.New(), PacketCreatedAt: start.Add(30 * time.Minute), MaxValue: 10},
		{PacketID: uuid.New(), PacketCreatedAt: start.Add(45 * time.Minute), MaxValue: 20},
	}

	mockRepo.On("GetMaxValuesByTimeRange", mock.Anything, start, end).
		Return(expectedData, nil)

	result, err := service.GetMaxValuesByTimeRange(context.Background(), start, end)
	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	mockRepo.AssertExpectations(t)
}

func TestDataService_GetMaxValuesByTimeRange_InvalidRange(t *testing.T) {
	mockRepo := new(MockRepository)
	logger, _ := zap.NewDevelopment()
	service := NewDataService(mockRepo, logger)

	end := time.Now().Add(-time.Hour)
	start := time.Now()

	result, err := service.GetMaxValuesByTimeRange(context.Background(), start, end)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end time must be after start time")
}
