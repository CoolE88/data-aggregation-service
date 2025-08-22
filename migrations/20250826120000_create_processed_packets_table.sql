-- +goose Up
CREATE TABLE IF NOT EXISTS processed_packets(
    packet_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    packet_created_at TIMESTAMPTZ NOT NULL,
    max_value INTEGER NOT NULL,
    PRIMARY KEY (packet_id, created_at)
) PARTITION BY RANGE (created_at);

-- +goose Down
DROP TABLE IF EXISTS processed_packets CASCADE;