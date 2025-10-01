package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seata-team/seata-go-client"
)

func tccExample() {
	// Create Seata client
	config := seata.DefaultConfig()
	config.HTTPEndpoint = "http://localhost:36789"

	client := seata.NewClient(config)
	defer client.Close()

	// Create TCC manager
	tccManager := seata.NewTCCManager(client)

	// Define TCC workflow
	workflow := seata.CreateTCCWorkflow([]seata.TCCStep{
		{
			BranchID: "order-service",
			Try:      "http://order-service:8080/api/orders/try",
			Confirm:  "http://order-service:8080/api/orders/confirm",
			Cancel:   "http://order-service:8080/api/orders/cancel",
		},
		{
			BranchID: "payment-service",
			Try:      "http://payment-service:8080/api/payments/try",
			Confirm:  "http://payment-service:8080/api/payments/confirm",
			Cancel:   "http://payment-service:8080/api/payments/cancel",
		},
		{
			BranchID: "inventory-service",
			Try:      "http://inventory-service:8080/api/inventory/try",
			Confirm:  "http://inventory-service:8080/api/inventory/confirm",
			Cancel:   "http://inventory-service:8080/api/inventory/cancel",
		},
	})

	// Validate workflow
	if err := workflow.Validate(); err != nil {
		log.Fatalf("Invalid workflow: %v", err)
	}

	// Create execution options
	options := seata.DefaultExecutionOptions()
	options.Timeout = 60 * time.Second
	options.ParallelBranches = true
	options.MaxConcurrency = 5

	// Execute TCC workflow
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)

	fmt.Println("Starting TCC workflow...")
	err := tccManager.ExecuteTCC(ctx, workflow, payload, options)
	if err != nil {
		log.Fatalf("TCC execution failed: %v", err)
	}

	fmt.Println("TCC workflow completed successfully!")
}

// Example with barrier pattern for idempotency
func tccWithBarrier() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	// Create TCC manager
	tccManager := seata.NewTCCManager(client)

	// Define TCC workflow
	workflow := seata.CreateTCCWorkflow([]seata.TCCStep{
		{
			BranchID: "order-service",
			Try:      "http://order-service:8080/api/orders/try",
			Confirm:  "http://order-service:8080/api/orders/confirm",
			Cancel:   "http://order-service:8080/api/orders/cancel",
		},
		{
			BranchID: "payment-service",
			Try:      "http://payment-service:8080/api/payments/try",
			Confirm:  "http://payment-service:8080/api/payments/confirm",
			Cancel:   "http://payment-service:8080/api/payments/cancel",
		},
	})

	// Execute TCC workflow with barrier pattern
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	barrierID := "barrier-12345"
	options := seata.DefaultExecutionOptions()

	fmt.Println("Starting TCC workflow with barrier...")
	err := tccManager.ExecuteTCCWithBarrier(ctx, workflow, payload, barrierID, options)
	if err != nil {
		log.Fatalf("TCC execution with barrier failed: %v", err)
	}

	fmt.Println("TCC workflow with barrier completed successfully!")
}

// Example of manual TCC transaction management
func manualTCCTransaction() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)

	// Start TCC transaction
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}

	fmt.Printf("Started TCC transaction: %s\n", tx.GetGID())

	// Try phase
	fmt.Println("Executing try phase...")
	err = tx.Try(ctx, "order-service", "http://order-service:8080/api/orders/try", payload)
	if err != nil {
		log.Fatalf("Try phase failed: %v", err)
	}

	err = tx.Try(ctx, "payment-service", "http://payment-service:8080/api/payments/try", payload)
	if err != nil {
		log.Fatalf("Try phase failed: %v", err)
	}

	// Confirm phase
	fmt.Println("Executing confirm phase...")
	err = tx.Confirm(ctx, "order-service")
	if err != nil {
		log.Fatalf("Confirm phase failed: %v", err)
	}

	err = tx.Confirm(ctx, "payment-service")
	if err != nil {
		log.Fatalf("Confirm phase failed: %v", err)
	}

	fmt.Println("TCC transaction completed successfully!")
}
