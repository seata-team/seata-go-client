package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seata-team/seata-go-client"
)

func grpcExample() {
	fmt.Println("=== Seata Go Client gRPC Example ===")

	// Create client with gRPC configuration
	config := &seata.Config{
		HTTPEndpoint:    "http://localhost:36789",
		GrpcEndpoint:    "localhost:36790",
		RequestTimeout:  30 * time.Second,
		MaxConnsPerHost: 100,
	}

	client := seata.NewClient(config)
	defer client.Close()

	ctx := context.Background()

	// Test gRPC transaction creation
	fmt.Println("1. Testing gRPC transaction creation...")
	payload := []byte(`{"order_id": "grpc-12345", "amount": 200.00}`)

	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Printf("❌ Failed to start transaction: %v", err)
		return
	}

	fmt.Printf("✅ Transaction created via gRPC: %s\n", tx.GetGID())

	// Test adding branches
	fmt.Println("\n2. Testing branch addition...")
    branches := []struct {
		id     string
		action string
	}{
        {"order-service-grpc", "http://127.0.0.1:18080/ok"},
        {"payment-service-grpc", "http://127.0.0.1:18080/ok"},
	}

	for _, branch := range branches {
		err = tx.AddBranch(ctx, branch.id, branch.action)
		if err != nil {
			log.Printf("❌ Failed to add branch %s: %v", branch.id, err)
		} else {
			fmt.Printf("✅ Added branch: %s\n", branch.id)
		}
	}

	// Test transaction submission
	fmt.Println("\n3. Testing transaction submission...")
	err = tx.Submit(ctx)
	if err != nil {
		log.Printf("❌ Failed to submit transaction: %v", err)
		return
	}

	fmt.Println("✅ Transaction submitted successfully!")

	// Monitor transaction
	fmt.Println("\n4. Monitoring transaction...")
	for i := 0; i < 10; i++ {
		info, err := client.GetTransaction(ctx, tx.GetGID())
		if err != nil {
			log.Printf("Failed to get transaction info: %v", err)
			break
		}

		fmt.Printf("Status: %s, Branches: %d\n", info.Status, len(info.Branches))

		switch info.Status {
		case seata.StatusCommitted:
			fmt.Println("✅ Transaction completed successfully!")
			return
		case seata.StatusAborted:
			fmt.Println("❌ Transaction was aborted!")
			return
		case seata.StatusSubmitted:
			time.Sleep(1 * time.Second)
			continue
		default:
			fmt.Printf("Unknown status: %s\n", info.Status)
			time.Sleep(1 * time.Second)
			continue
		}
	}

	fmt.Println("⏰ Transaction monitoring timeout!")
}
