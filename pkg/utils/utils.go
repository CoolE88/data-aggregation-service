package utils

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func NewUUID() uuid.UUID {
	return uuid.New()
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// GenerateRandomPayload генерит рандомные инты
func GenerateRandomPayload(size int) []int {
	payload := make([]int, size)
	for i := 0; i < size; i++ {
		payload[i] = rand.Intn(100)
	}
	return payload
}

type TimeGenerator struct {
	minTime time.Time
	maxTime time.Time
}

func NewTimeGenerator(min, max time.Time) *TimeGenerator {
	return &TimeGenerator{
		minTime: min,
		maxTime: max,
	}
}

func (tg *TimeGenerator) Generate() time.Time {
	delta := tg.maxTime.Sub(tg.minTime)
	randomDuration := time.Duration(rand.Int63n(int64(delta)))
	return tg.minTime.Add(randomDuration)
}

// DefaultTimeGenerator генератор за последние 30 дней
func DefaultTimeGenerator() *TimeGenerator {
	now := time.Now()
	return NewTimeGenerator(now.AddDate(0, 0, -30), now)
}

// PartitionedTimeGenerator создаёт генератор, который генерирует даты
// только в диапазоне существующих партиций: с 2025-08-01 по 2026-09-01 UTC
func PartitionedTimeGenerator() *TimeGenerator {
	loc := time.UTC
	min := time.Date(2025, 8, 1, 0, 0, 0, 0, loc)
	max := time.Date(2026, 9, 1, 0, 0, 0, 0, loc)
	return NewTimeGenerator(min, max)
}
