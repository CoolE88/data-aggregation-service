-- +goose Up
CREATE TABLE IF NOT EXISTS processed_packets_08_2025 PARTITION OF processed_packets
    FOR VALUES FROM ('2025-08-01 00:00:00+00') TO ('2025-09-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_08_2025_packet_created_at ON processed_packets_08_2025 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_09_2025 PARTITION OF processed_packets
    FOR VALUES FROM ('2025-09-01 00:00:00+00') TO ('2025-10-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_09_2025_packet_created_at ON processed_packets_09_2025 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_10_2025 PARTITION OF processed_packets
    FOR VALUES FROM ('2025-10-01 00:00:00+00') TO ('2025-11-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_10_2025_packet_created_at ON processed_packets_10_2025 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_11_2025 PARTITION OF processed_packets
    FOR VALUES FROM ('2025-11-01 00:00:00+00') TO ('2025-12-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_11_2025_packet_created_at ON processed_packets_11_2025 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_12_2025 PARTITION OF processed_packets
    FOR VALUES FROM ('2025-12-01 00:00:00+00') TO ('2026-01-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_12_2025_packet_created_at ON processed_packets_12_2025 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_01_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_01_2026_packet_created_at ON processed_packets_01_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_02_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-02-01 00:00:00+00') TO ('2026-03-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_02_2026_packet_created_at ON processed_packets_02_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_03_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-03-01 00:00:00+00') TO ('2026-04-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_03_2026_packet_created_at ON processed_packets_03_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_04_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-04-01 00:00:00+00') TO ('2026-05-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_04_2026_packet_created_at ON processed_packets_04_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_05_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-05-01 00:00:00+00') TO ('2026-06-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_05_2026_packet_created_at ON processed_packets_05_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_06_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-06-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_06_2026_packet_created_at ON processed_packets_06_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_07_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_07_2026_packet_created_at ON processed_packets_07_2026 (packet_created_at);

CREATE TABLE IF NOT EXISTS processed_packets_08_2026 PARTITION OF processed_packets
    FOR VALUES FROM ('2026-08-01 00:00:00+00') TO ('2026-09-01 00:00:00+00');

CREATE INDEX IF NOT EXISTS idx_processed_packets_08_2026_packet_created_at ON processed_packets_08_2026 (packet_created_at);

-- +goose Down
DROP TABLE IF EXISTS processed_packets_08_2025;
DROP TABLE IF EXISTS processed_packets_09_2025;
DROP TABLE IF EXISTS processed_packets_10_2025;
DROP TABLE IF EXISTS processed_packets_11_2025;
DROP TABLE IF EXISTS processed_packets_12_2025;
DROP TABLE IF EXISTS processed_packets_01_2026;
DROP TABLE IF EXISTS processed_packets_02_2026;
DROP TABLE IF EXISTS processed_packets_03_2026;
DROP TABLE IF EXISTS processed_packets_04_2026;
DROP TABLE IF EXISTS processed_packets_05_2026;
DROP TABLE IF EXISTS processed_packets_06_2026;
DROP TABLE IF EXISTS processed_packets_07_2026;
DROP TABLE IF EXISTS processed_packets_08_2026;
