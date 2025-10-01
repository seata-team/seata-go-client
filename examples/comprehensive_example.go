package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/seata-go-client"
)

func comprehensiveExample() {
	fmt.Println("=== Seata Go Client Comprehensive Example ===")

	// Create client with custom configuration
	config := &seata.Config{
		HTTPEndpoint:    "http://localhost:36789",
		GrpcEndpoint:    "localhost:36790",
		RequestTimeout:  30 * time.Second,
		RetryInterval:   1 * time.Second,
		MaxRetries:      3,
		MaxIdleConns:    100,
		MaxConnsPerHost: 100,
	}

	client := seata.NewClient(config)
	defer client.Close()

	ctx := context.Background()

	// 1. Health Check
	fmt.Println("\n1. Checking server health...")
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Printf("Server status: %s\n", health.Status)
	}

	// 2. Basic Transaction
	fmt.Println("\n2. Creating basic transaction...")
	payload := []byte(`{"order_id": "12345", "amount": 100.00, "user_id": "user123"}`)

	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return
	}

	fmt.Printf("Transaction ID: %s\n", tx.GetGID())

	// 3. Add Branches
	fmt.Println("\n3. Adding branches...")
	branches := []struct {
		id     string
		action string
	}{
		{"order-service", "http://order-service:8080/api/orders"},
		{"payment-service", "http://payment-service:8080/api/payments"},
		{"inventory-service", "http://inventory-service:8080/api/inventory"},
		{"notification-service", "http://notification-service:8080/api/notifications"},
	}

	for _, branch := range branches {
		err = tx.AddBranch(ctx, branch.id, branch.action)
		if err != nil {
			log.Printf("Failed to add branch %s: %v", branch.id, err)
		} else {
			fmt.Printf("Added branch: %s\n", branch.id)
		}
	}

	// 4. Submit Transaction
	fmt.Println("\n4. Submitting transaction...")
	err = tx.Submit(ctx)
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
		return
	}

	// 5. Monitor Transaction
	fmt.Println("\n5. Monitoring transaction...")
	monitorTransaction(ctx, client, tx.GetGID())

	// 6. Saga Workflow Example
	fmt.Println("\n6. Saga Workflow Example...")
	sagaWorkflowExample(client)

	// 7. TCC Workflow Example
	fmt.Println("\n7. TCC Workflow Example...")
	tccWorkflowExample(client)

	// 8. Error Handling Example
	fmt.Println("\n8. Error Handling Example...")
	errorHandlingExampleComprehensive(client)

	// 9. Retry and Circuit Breaker Example
	fmt.Println("\n9. Retry and Circuit Breaker Example...")
	retryAndCircuitBreakerExample()

	// 10. Query Operations
	fmt.Println("\n10. Query Operations...")
	queryOperationsExample(client)
}

func monitorTransaction(ctx context.Context, client *seata.Client, gid string) {
	fmt.Printf("Monitoring transaction: %s\n", gid)

	for i := 0; i < 10; i++ { // Monitor for up to 10 seconds
		info, err := client.GetTransaction(ctx, gid)
		if err != nil {
			log.Printf("Failed to get transaction info: %v", err)
			break
		}

		fmt.Printf("Status: %s, Branches: %d\n", info.Status, len(info.Branches))

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

	fmt.Println("Transaction monitoring timeout!")
}

func sagaWorkflowExample(client *seata.Client) {
	// Create Saga manager
	sagaManager := seata.NewSagaManager(client)

	// Define workflow
	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID:   "create-order",
			Action:     "http://order-service:8080/api/orders",
			Compensate: "http://order-service:8080/api/orders/compensate",
		},
		{
			BranchID:   "process-payment",
			Action:     "http://payment-service:8080/api/payments",
			Compensate: "http://payment-service:8080/api/payments/compensate",
		},
		{
			BranchID:   "update-inventory",
			Action:     "http://inventory-service:8080/api/inventory",
			Compensate: "http://inventory-service:8080/api/inventory/compensate",
		},
	})

	// Validate workflow
	if err := workflow.Validate(); err != nil {
		log.Printf("Invalid workflow: %v", err)
		return
	}

	// Custom compensation function
	compensationFunc := func(ctx context.Context, failedStep *seata.SagaStep) error {
		fmt.Printf("Executing custom compensation for branch: %s\n", failedStep.BranchID)
		// Implement custom compensation logic here
		return nil
	}

	// Execute workflow
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	options := seata.DefaultExecutionOptions()
	options.Timeout = 60 * time.Second

	fmt.Println("Executing Saga workflow with custom compensation...")
	err := sagaManager.ExecuteSagaWithCompensation(ctx, workflow, payload, compensationFunc, options)
	if err != nil {
		log.Printf("Saga execution failed: %v", err)
	} else {
		fmt.Println("Saga workflow completed successfully!")
	}
}

func tccWorkflowExample(client *seata.Client) {
	// Create TCC manager
	tccManager := seata.NewTCCManager(client)

	// Define workflow
	workflow := seata.CreateTCCWorkflow([]seata.TCCStep{
		{
			BranchID: "reserve-order",
			Try:      "http://order-service:8080/api/orders/try",
			Confirm:  "http://order-service:8080/api/orders/confirm",
			Cancel:   "http://order-service:8080/api/orders/cancel",
		},
		{
			BranchID: "reserve-payment",
			Try:      "http://payment-service:8080/api/payments/try",
			Confirm:  "http://payment-service:8080/api/payments/confirm",
			Cancel:   "http://payment-service:8080/api/payments/cancel",
		},
		{
			BranchID: "reserve-inventory",
			Try:      "http://inventory-service:8080/api/inventory/try",
			Confirm:  "http://inventory-service:8080/api/inventory/confirm",
			Cancel:   "http://inventory-service:8080/api/inventory/cancel",
		},
	})

	// Validate workflow
	if err := workflow.Validate(); err != nil {
		log.Printf("Invalid workflow: %v", err)
		return
	}

	// Execute workflow
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	barrierID := "barrier-12345"
	options := seata.DefaultExecutionOptions()
	options.Timeout = 60 * time.Second
	options.ParallelBranches = true
	options.MaxConcurrency = 5

	fmt.Println("Executing TCC workflow with barrier pattern...")
	err := tccManager.ExecuteTCCWithBarrier(ctx, workflow, payload, barrierID, options)
	if err != nil {
		log.Printf("TCC execution failed: %v", err)
	} else {
		fmt.Println("TCC workflow completed successfully!")
	}
}

func errorHandlingExampleComprehensive(client *seata.Client) {
	ctx := context.Background()

	// Test error handling
	fmt.Println("Testing error handling...")

	// Try to get non-existent transaction
	_, err := client.GetTransaction(ctx, "non-existent-id")
	if err != nil {
		fmt.Printf("Expected error for non-existent transaction: %v\n", err)
	}

	// Try to start transaction with invalid server
	invalidConfig := seata.DefaultConfig()
	invalidConfig.HTTPEndpoint = "http://invalid-server:9999"
	invalidClient := seata.NewClient(invalidConfig)
	defer invalidClient.Close()

	_, err = invalidClient.StartTransaction(ctx, seata.ModeSaga, []byte("test"))
	if err != nil {
		fmt.Printf("Expected connection error: %v\n", err)
	}
}

func retryAndCircuitBreakerExample() {
	// Retry configuration
	retryConfig := &seata.RetryConfig{
		MaxRetries:    5,
		RetryInterval: 2 * time.Second,
		BackoffFactor: 2.0,
	}

	retryManager := seata.NewRetryManager(retryConfig)

	// Circuit breaker configuration
	circuitBreakerConfig := &seata.CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  10 * time.Second,
		HalfOpenMaxCalls: 2,
	}

	circuitBreaker := seata.NewCircuitBreaker(circuitBreakerConfig)

	// Test retry mechanism
	fmt.Println("Testing retry mechanism...")
	ctx := context.Background()

	attemptCount := 0
	err := retryManager.ExecuteWithRetry(ctx, func() error {
		attemptCount++
		fmt.Printf("Retry attempt %d\n", attemptCount)

		if attemptCount < 3 {
			return fmt.Errorf("simulated error")
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Retry failed: %v\n", err)
	} else {
		fmt.Printf("Retry succeeded after %d attempts\n", attemptCount)
	}

	// Test circuit breaker
	fmt.Println("Testing circuit breaker...")

	// Simulate failures to open circuit
	for i := 0; i < 5; i++ {
		err := circuitBreaker.Execute(func() error {
			return fmt.Errorf("simulated failure")
		})

		if err != nil {
			fmt.Printf("Circuit breaker execution %d failed: %v\n", i+1, err)
		}
	}

	fmt.Printf("Circuit breaker state: %v\n", circuitBreaker.GetState())
}

func queryOperationsExample(client *seata.Client) {
	ctx := context.Background()

	// List transactions
	fmt.Println("Listing transactions...")
	transactions, err := client.ListTransactions(ctx, 10, 0, "")
	if err != nil {
		log.Printf("Failed to list transactions: %v", err)
		return
	}

	fmt.Printf("Found %d transactions:\n", len(transactions))
	for i, tx := range transactions {
		if i >= 3 { // Limit output
			fmt.Printf("... and %d more\n", len(transactions)-3)
			break
		}
		fmt.Printf("- ID: %s, Status: %s, Mode: %s\n", tx.GID, tx.Status, tx.Mode)
	}

	// Get metrics
	fmt.Println("\nGetting server metrics...")
	metrics, err := client.Metrics(ctx)
	if err != nil {
		log.Printf("Failed to get metrics: %v", err)
	} else {
		fmt.Printf("Metrics (first 200 chars): %s...\n", metrics[:min(200, len(metrics))])
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
