ifneq (,$(wildcard .env))
  include .env
  export
endif

PROTO_DIR=api/proto
GEN_DIR=gen/go
MIGRATE_DIR=migrations

DB_URL = "user=$(PGUSER) password=$(PGPASSWORD) dbname=$(PGDATABASE) host=$(PGHOST) port=$(PGPORT) sslmode=$(PGSSLMODE)"

.PHONY: proto build test test-coverage lint migrate clean

proto:
	mkdir -p $(GEN_DIR)
	protoc --go_out=$(GEN_DIR) \
	       --go_opt=paths=source_relative \
	       --go-grpc_out=$(GEN_DIR) \
	       --go-grpc_opt=paths=source_relative \
	       -I $(PROTO_DIR) \
	       $(PROTO_DIR)/aggregator/v1/aggregator.proto
	@echo "Protobuf code generated in $(GEN_DIR)"

build:
	go build -o data-aggregation-service ./cmd/main.go

test:
	go test -v ./internal/service/... ./internal/grpc/... ./internal/http/... ./internal/aggregator/... ./pkg/utils/...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf $(GEN_DIR)
	rm -f data-aggregation-service
	rm -f coverage.out

migrate:
	goose -dir $(MIGRATE_DIR) postgres "$(DB_URL)" up
