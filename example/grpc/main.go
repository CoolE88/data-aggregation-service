package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/CoolE88/data-aggregation-service/gen/go/aggregator/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.Dial("localhost:9090", //nolint:staticcheck // Will be supported throughout 1.x
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	client := pb.NewDataAggregationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Тест 1: GetMaxValuesByPeriod
	fmt.Println("=== Test 1: GetMaxValuesByPeriod ===")
	testGetMaxValuesByPeriod(ctx, client)

	// Тест 2: GetMaxValueByID
	fmt.Println("\n=== Test 2: GetMaxValueByID ===")
	testGetMaxValueByID(ctx, client)

	// Тест 3: Ошибки валидации
	fmt.Println("\n=== Test 3: Validation Errors ===")
	testValidationErrors(ctx, client)
}

func testGetMaxValuesByPeriod(ctx context.Context, client pb.DataAggregationServiceClient) {
	startTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)

	req := &pb.TimePeriod{
		StartTime: startTime,
		EndTime:   endTime,
	}

	resp, err := client.GetMaxValuesByPeriod(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("gRPC error: %s (code: %s)", st.Message(), st.Code())
		} else {
			log.Printf("Error: %v", err)
		}
		return
	}

	fmt.Printf("Found %d max values:\n", len(resp.MaxValues))
	for i, val := range resp.MaxValues {
		fmt.Printf("%d. ID: %s, MaxValue: %d\n", i+1, val.Id, val.MaxValue)
	}
}

func testGetMaxValueByID(ctx context.Context, client pb.DataAggregationServiceClient) {
	packetID := "123e4567-e89b-12d3-a456-426614174000"

	req := &pb.PackageID{Id: packetID}

	resp, err := client.GetMaxValueByID(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				fmt.Printf("Packet with ID %s not found\n", packetID)
			} else {
				log.Printf("gRPC error: %s (code: %s)", st.Message(), st.Code())
			}
		} else {
			log.Printf("Error: %v", err)
		}
		return
	}

	fmt.Printf("Found packet: ID=%s, MaxValue=%d\n", resp.Id, resp.MaxValue)
}

func testValidationErrors(ctx context.Context, client pb.DataAggregationServiceClient) {
	// Тест пустого запроса
	fmt.Println("Testing empty time period...")
	_, err := client.GetMaxValuesByPeriod(ctx, &pb.TimePeriod{})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			fmt.Printf("Expected error: %s (code: %s)\n", st.Message(), st.Code())
		}
	}

	// Тест невалидного формата времени
	fmt.Println("Testing invalid time format...")
	_, err = client.GetMaxValuesByPeriod(ctx, &pb.TimePeriod{
		StartTime: "invalid-date",
		EndTime:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			fmt.Printf("Expected error: %s (code: %s)\n", st.Message(), st.Code())
		}
	}

	// Тест пустого ID
	fmt.Println("Testing empty package ID...")
	_, err = client.GetMaxValueByID(ctx, &pb.PackageID{})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			fmt.Printf("Expected error: %s (code: %s)\n", st.Message(), st.Code())
		}
	}
}
