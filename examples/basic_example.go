package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seata-team/seata-go-client"
)

func basicExample() {
	// Create Seata client with default configuration
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()

	// Check server health
	fmt.Println("Checking server health...")
	health, err := client.Health(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("Server status: %s\n", health.Status)

	// Start a Saga transaction
	fmt.Println("Starting Saga transaction...")
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}

	fmt.Printf("Transaction ID: %s\n", tx.GetGID())

	// Add branches to the transaction
	fmt.Println("Adding branches...")
	err = tx.AddBranch(ctx, "order-service", "http://order-service:8080/api/orders")
	if err != nil {
		log.Fatalf("Failed to add order branch: %v", err)
	}

	err = tx.AddBranch(ctx, "payment-service", "http://payment-service:8080/api/payments")
	if err != nil {
		log.Fatalf("Failed to add payment branch: %v", err)
	}

	err = tx.AddBranch(ctx, "inventory-service", "http://inventory-service:8080/api/inventory")
	if err != nil {
		log.Fatalf("Failed to add inventory branch: %v", err)
	}

	// Submit the transaction
	fmt.Println("Submitting transaction...")
	err = tx.Submit(ctx)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	// Wait for completion
	fmt.Println("Waiting for transaction completion...")
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		info, err := tx.GetInfo(ctx)
		if err != nil {
			log.Fatalf("Failed to get transaction info: %v", err)
		}

		fmt.Printf("Transaction status: %s\n", info.Status)

		switch info.Status {
		case seata.StatusCommitted:
			fmt.Println("Transaction completed successfully!")
			return
		case seata.StatusAborted:
			fmt.Println("Transaction was aborted!")
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

	fmt.Println("Transaction timeout!")
}

// Example of TCC transaction
func tccExampleBasic() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)

	// Start TCC transaction
	fmt.Println("Starting TCC transaction...")
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("Failed to start TCC transaction: %v", err)
	}

	fmt.Printf("TCC Transaction ID: %s\n", tx.GetGID())

	// Try phase
	fmt.Println("Executing try phase...")
	err = tx.Try(ctx, "order-service", "http://order-service:8080/api/orders/try", payload)
	if err != nil {
		log.Fatalf("Try phase failed: %v", err)
	}

	// Confirm phase
	fmt.Println("Executing confirm phase...")
	err = tx.Confirm(ctx, "order-service")
	if err != nil {
		log.Fatalf("Confirm phase failed: %v", err)
	}

	fmt.Println("TCC transaction completed successfully!")
}

// Example of transaction querying
func queryExample() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()

	// List transactions
	fmt.Println("Listing transactions...")
	transactions, err := client.ListTransactions(ctx, 10, 0, "")
	if err != nil {
		log.Fatalf("Failed to list transactions: %v", err)
	}

	fmt.Printf("Found %d transactions:\n", len(transactions))
	for _, tx := range transactions {
		fmt.Printf("- ID: %s, Status: %s, Mode: %s\n", tx.GID, tx.Status, tx.Mode)
	}

	// Get specific transaction
	if len(transactions) > 0 {
		fmt.Printf("\nGetting transaction details for: %s\n", transactions[0].GID)
		txInfo, err := client.GetTransaction(ctx, transactions[0].GID)
		if err != nil {
			log.Fatalf("Failed to get transaction: %v", err)
		}

		fmt.Printf("Transaction details:\n")
		fmt.Printf("- ID: %s\n", txInfo.GID)
		fmt.Printf("- Status: %s\n", txInfo.Status)
		fmt.Printf("- Mode: %s\n", txInfo.Mode)
		fmt.Printf("- Branches: %d\n", len(txInfo.Branches))
		for _, branch := range txInfo.Branches {
			fmt.Printf("  - Branch %s: %s (%s)\n", branch.BranchID, branch.Action, branch.Status)
		}
	}
}

// Example of error handling
func errorHandlingExampleBasic() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()

	// Try to get a non-existent transaction
	fmt.Println("Trying to get non-existent transaction...")
	_, err := client.GetTransaction(ctx, "non-existent-id")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Try to start transaction with invalid server
	fmt.Println("Trying to connect to invalid server...")
	invalidConfig := seata.DefaultConfig()
	invalidConfig.HTTPEndpoint = "http://invalid-server:9999"
	invalidClient := seata.NewClient(invalidConfig)
	defer invalidClient.Close()

	_, err = invalidClient.StartTransaction(ctx, seata.ModeSaga, []byte("test"))
	if err != nil {
		fmt.Printf("Expected connection error: %v\n", err)
	}
}
