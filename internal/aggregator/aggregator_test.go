package aggregator

import (
	"context"
	"testing"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) ProcessPacket(ctx context.Context, packet *domain.DataPacket) error {
	args := m.Called(ctx, packet)
	return args.Error(0)
}

func TestAggregator_Start(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	aggregator := NewAggregator(mockService, 2, logger)

	packets := make(chan *domain.DataPacket, 3)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	packet1 := &domain.DataPacket{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Payload:   []int{1, 2, 3},
	}
	packet2 := &domain.DataPacket{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Payload:   []int{4, 5, 6},
	}

	mockService.On("ProcessPacket", mock.Anything, packet1).Return(nil)
	mockService.On("ProcessPacket", mock.Anything, packet2).Return(nil)

	go func() {
		packets <- packet1
		packets <- packet2
		close(packets)
	}()

	aggregator.Start(ctx, packets)

	// Ждём, пока все пакеты обработаются
	aggregator.Wait()

	mockService.AssertExpectations(t)
	mockService.AssertNumberOfCalls(t, "ProcessPacket", 2)
}

func TestAggregator_Stop(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	aggregator := NewAggregator(mockService, 1, logger)

	packets := make(chan *domain.DataPacket, 1)
	ctx := context.Background()

	aggregator.Start(ctx, packets)

	// Дать время воркеру стартовать
	time.Sleep(10 * time.Millisecond)

	// Останавливаем агрегатор
	aggregator.Stop()

	// Ждём, пока воркер завершится
	done := make(chan struct{})
	go func() {
		aggregator.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Успешно завершился
	case <-time.After(100 * time.Millisecond):
		t.Error("Worker did not stop after Stop()")
	}
}
