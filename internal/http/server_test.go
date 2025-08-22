package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/domain"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

func TestHTTPServer_HealthCheck(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	server := NewHTTPServer(":8080", mockService, logger)

	mockService.On("CheckDBConnection", mock.Anything).Return(nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	mockService.AssertExpectations(t)
}

func TestHTTPServer_GetMaxValuesByTimeRange(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	server := NewHTTPServer(":8080", mockService, logger)

	start := time.Now().Add(-time.Hour).UTC()
	end := time.Now().UTC()
	expectedData := []*domain.ProcessedData{
		{
			PacketID:        uuid.New(),
			PacketCreatedAt: start.Add(30 * time.Minute),
			MaxValue:        100,
		},
	}

	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	mockService.On("GetMaxValuesByTimeRange",
		mock.Anything,
		mock.MatchedBy(func(t time.Time) bool {
			return t.Truncate(time.Second).Equal(start.Truncate(time.Second))
		}),
		mock.MatchedBy(func(t time.Time) bool {
			return t.Truncate(time.Second).Equal(end.Truncate(time.Second))
		}),
	).Return(expectedData, nil)

	req := httptest.NewRequest(
		"GET",
		"/api/v1/max-values?start="+startStr+"&end="+endStr,
		nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/max-values", server.getMaxValuesByTimeRange).Methods("GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*domain.ProcessedData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, 100, response[0].MaxValue)

	mockService.AssertExpectations(t)
}

func TestHTTPServer_GetMaxValueByID(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	server := NewHTTPServer(":8080", mockService, logger)

	packetID := uuid.New()
	expectedData := &domain.ProcessedData{
		PacketID:        packetID,
		PacketCreatedAt: time.Now(),
		MaxValue:        42,
	}

	mockService.On("GetMaxValueByPacketID", mock.Anything, packetID.String()).
		Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/max-values/"+packetID.String(), nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/max-values/{id}", server.getMaxValueByID).Methods("GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.ProcessedData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 42, response.MaxValue)
	mockService.AssertExpectations(t)
}
