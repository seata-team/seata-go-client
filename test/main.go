package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seata-team/seata-go-client"
)

func main() {
	fmt.Println("=== Seata Go Client Test ===")
	
	// Create client
	client := seata.NewClientWithDefaults()
	defer client.Close()
	
	ctx := context.Background()
	
	// Test health check
	fmt.Println("1. Testing health check...")
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Printf("✅ Server status: %s\n", health.Status)
	}
	
	// Test transaction creation
	fmt.Println("\n2. Testing transaction creation...")
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Printf("❌ Failed to start transaction: %v", err)
		return
	}
	
	fmt.Printf("✅ Transaction created: %s\n", tx.GetGID())
	
	// Test adding branches
	fmt.Println("\n3. Testing branch addition...")
	branches := []struct {
		id     string
		action string
	}{
		{"order-service", "http://httpbin.org/status/200"},
		{"payment-service", "http://httpbin.org/status/201"},
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
	fmt.Println("\n4. Testing transaction submission...")
	err = tx.Submit(ctx)
	if err != nil {
		log.Printf("❌ Failed to submit transaction: %v", err)
		return
	}
	
	fmt.Println("✅ Transaction submitted successfully!")
	
	// Monitor transaction
	fmt.Println("\n5. Monitoring transaction...")
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