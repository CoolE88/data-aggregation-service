package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeGenerator_Generate(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	generator := NewTimeGenerator(start, end)

	for i := 0; i < 100; i++ {
		result := generator.Generate()
		assert.True(t, !result.Before(start) && !result.After(end))
	}
}

func TestGenerateRandomPayload(t *testing.T) {
	for size := 1; size <= 100; size++ {
		payload := GenerateRandomPayload(size)
		assert.Len(t, payload, size)

		for _, val := range payload {
			assert.True(t, val >= -1000 && val <= 1000)
		}
	}
}

func TestNewUUID(t *testing.T) {
	uuid := NewUUID()
	assert.NotEmpty(t, uuid.String())
}
