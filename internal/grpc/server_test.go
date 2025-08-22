package grpc

import (
	"context"
	"testing"
	"time"

	pb "github.com/CoolE88/data-aggregation-service/gen/go/aggregator/v1"
	"github.com/CoolE88/data-aggregation-service/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetMaxValueByPacketID(ctx context.Context, packetID string) (*domain.ProcessedData, error) {
	args := m.Called(ctx, packetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProcessedData), args.Error(1)
}

func (m *MockService) GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error) {
	args := m.Called(ctx, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ProcessedData), args.Error(1)
}

func (m *MockService) CheckDBConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestGRPCServer_GetMaxValuesByPeriod(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	server := NewGRPCServer(mockService, logger)

	start := time.Date(2025, 8, 27, 14, 58, 37, 0, time.UTC)
	end := time.Date(2025, 8, 27, 15, 58, 37, 0, time.UTC)

	expectedData := []*domain.ProcessedData{
		{PacketID: uuid.New(), MaxValue: 100},
	}

	mockService.On("GetMaxValuesByTimeRange",
		mock.Anything,
		start,
		end,
	).Return(expectedData, nil)

	req := &pb.TimePeriod{
		StartTime: start.Format(time.RFC3339),
		EndTime:   end.Format(time.RFC3339),
	}

	ctx := context.Background()
	resp, err := server.GetMaxValuesByPeriod(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, resp.MaxValues, 1)
	assert.Equal(t, int32(100), resp.MaxValues[0].MaxValue)

	mockService.AssertExpectations(t)
}

func TestGRPCServer_GetMaxValuesByPeriod_InvalidTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	server := &GRPCServer{service: mockService, logger: logger}

	req := &pb.TimePeriod{
		StartTime: "invalid-time",
		EndTime:   time.Now().Format(time.RFC3339),
	}

	response, err := server.GetMaxValuesByPeriod(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGRPCServer_GetMaxValueByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	server := &GRPCServer{service: mockService, logger: logger}

	packetID := uuid.New()
	expectedData := &domain.ProcessedData{
		PacketID:        packetID,
		PacketCreatedAt: time.Now(),
		MaxValue:        42,
	}

	mockService.On("GetMaxValueByPacketID", mock.Anything, packetID.String()).
		Return(expectedData, nil)

	req := &pb.PackageID{Id: packetID.String()}
	response, err := server.GetMaxValueByID(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int32(42), response.MaxValue)
	assert.Equal(t, packetID.String(), response.Id)
	mockService.AssertExpectations(t)
}

func TestGRPCServer_GetMaxValueByID_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	server := &GRPCServer{service: mockService, logger: logger}

	packetID := uuid.New()
	mockService.On("GetMaxValueByPacketID", mock.Anything, packetID.String()).
		Return((*domain.ProcessedData)(nil), nil)

	req := &pb.PackageID{Id: packetID.String()}
	response, err := server.GetMaxValueByID(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
