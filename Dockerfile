FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git bash protoc protobuf-dev make

# Устанавливаем Go плагины для protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Копируем зависимости и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Генерируем protobuf/grpc код
RUN protoc --go_out=. --go-grpc_out=. \
    --proto_path=./api/proto \
    api/proto/aggregator/v1/aggregator.proto

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o data-aggregation-service ./cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/data-aggregation-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080 9090

CMD ["./data-aggregation-service"]
