package benchmarks

import (
	"context"
	"testing"

	"github.com/seata-team/seata-go-client"
)

func BenchmarkClientCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := seata.NewClientWithDefaults()
		client.Close()
	}
}

func BenchmarkTransactionCreation(b *testing.B) {
	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail without a running server, but we can measure the overhead
		_, _ = client.StartTransaction(ctx, seata.ModeSaga, payload)
	}
}

func BenchmarkSagaWorkflowCreation(b *testing.B) {
	client := seata.NewClientWithDefaults()
	defer client.Close()

	sagaManager := seata.NewSagaManager(client)

	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID:   "step1",
			Action:     "http://example.com/step1",
			Compensate: "http://example.com/step1/compensate",
		},
		{
			BranchID:   "step2",
			Action:     "http://example.com/step2",
			Compensate: "http://example.com/step2/compensate",
		},
	})

	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	options := seata.DefaultExecutionOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail without a running server, but we can measure the overhead
		_ = sagaManager.ExecuteSaga(ctx, workflow, payload, options)
	}
}

func BenchmarkTCCWorkflowCreation(b *testing.B) {
	client := seata.NewClientWithDefaults()
	defer client.Close()

	tccManager := seata.NewTCCManager(client)

	workflow := seata.CreateTCCWorkflow([]seata.TCCStep{
		{
			BranchID: "step1",
			Try:      "http://example.com/step1/try",
			Confirm:  "http://example.com/step1/confirm",
			Cancel:   "http://example.com/step1/cancel",
		},
		{
			BranchID: "step2",
			Try:      "http://example.com/step2/try",
			Confirm:  "http://example.com/step2/confirm",
			Cancel:   "http://example.com/step2/cancel",
		},
	})

	ctx := context.Background()
	payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
	options := seata.DefaultExecutionOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail without a running server, but we can measure the overhead
		_ = tccManager.ExecuteTCC(ctx, workflow, payload, options)
	}
}

func BenchmarkRetryManager(b *testing.B) {
	config := seata.DefaultRetryConfig()
	retryManager := seata.NewRetryManager(config)

	ctx := context.Background()
	operation := func() error {
		return nil // Always succeed for this benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = retryManager.ExecuteWithRetry(ctx, operation)
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	config := seata.DefaultCircuitBreakerConfig()
	circuitBreaker := seata.NewCircuitBreaker(config)

	operation := func() error {
		return nil // Always succeed for this benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = circuitBreaker.Execute(operation)
	}
}

func BenchmarkWorkflowValidation(b *testing.B) {
	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID:   "step1",
			Action:     "http://example.com/step1",
			Compensate: "http://example.com/step1/compensate",
		},
		{
			BranchID:   "step2",
			Action:     "http://example.com/step2",
			Compensate: "http://example.com/step2/compensate",
		},
		{
			BranchID:   "step3",
			Action:     "http://example.com/step3",
			Compensate: "http://example.com/step3/compensate",
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = workflow.Validate()
	}
}

func BenchmarkConfigCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = seata.DefaultConfig()
		_ = seata.DefaultRetryConfig()
		_ = seata.DefaultCircuitBreakerConfig()
		_ = seata.DefaultExecutionOptions()
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		client := seata.NewClientWithDefaults()

		// Create workflow
		workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
			{
				BranchID:   "step1",
				Action:     "http://example.com/step1",
				Compensate: "http://example.com/step1/compensate",
			},
		})

		// Validate workflow
		_ = workflow.Validate()

		// Create managers
		_ = seata.NewSagaManager(client)
		_ = seata.NewTCCManager(client)
		_ = seata.NewRetryManager(seata.DefaultRetryConfig())
		_ = seata.NewCircuitBreaker(seata.DefaultCircuitBreakerConfig())

		client.Close()
	}
}

// Concurrent access benchmarks
func BenchmarkConcurrentClientCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client := seata.NewClientWithDefaults()
			client.Close()
		}
	})
}

func BenchmarkConcurrentWorkflowValidation(b *testing.B) {
	workflow := seata.CreateSagaWorkflow([]seata.SagaStep{
		{
			BranchID:   "step1",
			Action:     "http://example.com/step1",
			Compensate: "http://example.com/step1/compensate",
		},
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = workflow.Validate()
		}
	})
}
