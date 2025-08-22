package domain

import (
	"time"

	"github.com/google/uuid"
)

// DataPacket представляет входящий пакет данных
type DataPacket struct {
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   []int     `json:"payload"`
}

// ProcessedData представляет обработанные данные
type ProcessedData struct {
	PacketID        uuid.UUID `json:"packet_id" db:"packet_id"`
	PacketCreatedAt time.Time `json:"packet_created_at" db:"packet_created_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	MaxValue        int       `json:"max_value" db:"max_value"`
}
