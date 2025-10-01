package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/seata-go-client"
)

func sagaExample() {
	// Create Seata client
	config := seata.DefaultConfig()
	config.HTTPEndpoint = "http://localhost:36789"

	client := seata.NewClient(config)
	defer client.Close()

	// Create Saga manager
	sagaManager := seata.NewSagaManager(client)

	// Define Saga workflow
	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID:   "order-service",
			Action:     "http://order-service:8080/api/orders",
			Compensate: "http://order-service:8080/api/orders/compensate",
		},
		{
			BranchID:   "payment-service",
			Action:     "http://payment-service:8080/api/payments",
			Compensate: "http://payment-service:8080/api/payments/compensate",
		},
		{
			BranchID:   "inventory-service",
			Action:     "http://inventory-service:8080/api/inventory",
			Compensate: "http://inventory-service:8080/api/inventory/compensate",
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

	// Execute Saga workflow
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)

	fmt.Println("Starting Saga workflow...")
	err := sagaManager.ExecuteSaga(ctx, workflow, payload, options)
	if err != nil {
		log.Fatalf("Saga execution failed: %v", err)
	}

	fmt.Println("Saga workflow completed successfully!")
}

// Example with custom compensation
func sagaWithCustomCompensation() {
	// Create Seata client
	client := seata.NewClientWithDefaults()
	defer client.Close()

	// Create Saga manager
	sagaManager := seata.NewSagaManager(client)

	// Define Saga workflow
	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID: "order-service",
			Action:   "http://order-service:8080/api/orders",
		},
		{
			BranchID: "payment-service",
			Action:   "http://payment-service:8080/api/payments",
		},
	})

	// Custom compensation function
	compensationFunc := func(ctx context.Context, failedStep *seata.SagaStep) error {
		fmt.Printf("Executing custom compensation for branch: %s\n", failedStep.BranchID)

		// Implement custom compensation logic here
		// For example, send notification, update database, etc.

		return nil
	}

	// Execute Saga with custom compensation
	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	options := seata.DefaultExecutionOptions()

	err := sagaManager.ExecuteSagaWithCompensation(ctx, workflow, payload, compensationFunc, options)
	if err != nil {
		log.Fatalf("Saga execution with custom compensation failed: %v", err)
	}

	fmt.Println("Saga workflow with custom compensation completed!")
}
