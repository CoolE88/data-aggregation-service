package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/CoolE88/data-aggregation-service/internal/config"
	"github.com/CoolE88/data-aggregation-service/internal/domain"
	"github.com/CoolE88/data-aggregation-service/internal/metrics"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepository(ctx context.Context, dbConfig config.DBConfig, logger *zap.Logger) (*PostgresRepository, error) {
	// Конфигурация пула
	config, err := pgxpool.ParseConfig(dbConfig.DBSource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Настройка пула
	config.MaxConns = int32(dbConfig.MaxDBConnections)
	config.MinConns = int32(dbConfig.MinDBConnections)
	config.MaxConnLifetime = dbConfig.MaxConnLifetime
	config.MaxConnIdleTime = dbConfig.MaxConnIdleTime

	// Создание пула
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Запуск горутины для мониторинга соединений
	go monitorConnections(ctx, pool, logger)

	return &PostgresRepository{
		pool:   pool,
		logger: logger,
	}, nil
}

// monitorConnections периодически обновляет метрики соединений и завершается при отмене ctx
func monitorConnections(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping monitorConnections goroutine due to context cancellation")
			return
		case <-ticker.C:
			stats := pool.Stat()
			metrics.DBActiveConnections.Set(float64(stats.AcquiredConns()))
			metrics.DBIdleConnections.Set(float64(stats.IdleConns()))

			logger.Debug("Database connection stats",
				zap.Int("acquired", int(stats.AcquiredConns())),
				zap.Int("idle", int(stats.IdleConns())),
				zap.Int("max", int(stats.MaxConns())),
			)
		}
	}
}

func (r *PostgresRepository) SaveProcessedData(ctx context.Context, data *domain.ProcessedData) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("save_processed_data").Observe(time.Since(start).Seconds())
	}()

	query := "INSERT INTO processed_packets (packet_id, packet_created_at, max_value, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (packet_id, created_at) DO NOTHING RETURNING packet_id"

	var insertedID uuid.UUID
	err := r.pool.QueryRow(ctx, query,
		data.PacketID,
		data.PacketCreatedAt,
		data.MaxValue,
		data.CreatedAt,
	).Scan(&insertedID)

	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to save processed data: %w", err)
	}

	if err == pgx.ErrNoRows {
		r.logger.Debug("duplicate packet ignored", zap.String("packet_id", data.PacketID.String()))
	}

	return nil
}

func (r *PostgresRepository) GetMaxValueByPacketID(ctx context.Context, packetID uuid.UUID) (*domain.ProcessedData, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("get_max_value_by_packet_id").Observe(time.Since(start).Seconds())
	}()

	query := "SELECT packet_id, packet_created_at, max_value, created_at FROM processed_packets WHERE packet_id = $1"

	var data domain.ProcessedData
	err := r.pool.QueryRow(ctx, query, packetID).Scan(
		&data.PacketID,
		&data.PacketCreatedAt,
		&data.MaxValue,
		&data.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get processed data: %w", err)
	}

	return &data, nil
}

func (r *PostgresRepository) GetMaxValuesByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.ProcessedData, error) {
	startTime := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("get_max_values_by_time_range").Observe(time.Since(startTime).Seconds())
	}()

	query := "SELECT packet_id, packet_created_at, max_value, created_at FROM processed_packets WHERE created_at >= $1 AND created_at < $2 ORDER BY packet_created_at"

	rows, err := r.pool.Query(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query processed data: %w", err)
	}
	defer rows.Close()

	var results []*domain.ProcessedData
	for rows.Next() {
		var data domain.ProcessedData
		err := rows.Scan(
			&data.PacketID,
			&data.PacketCreatedAt,
			&data.MaxValue,
			&data.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, &data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (r *PostgresRepository) HealthCheck(ctx context.Context) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.DBQueryDuration.WithLabelValues("health_check").Observe(duration)
	}()

	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
