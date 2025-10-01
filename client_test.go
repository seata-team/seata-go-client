package seata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := DefaultConfig()
	client := NewClient(config)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.config)
	assert.Equal(t, config, client.config)
}

func TestNewClientWithDefaults(t *testing.T) {
	client := NewClientWithDefaults()

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.config)
	assert.Equal(t, "http://localhost:36789", client.config.HTTPEndpoint)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "http://localhost:36789", config.HTTPEndpoint)
	assert.Equal(t, "localhost:36790", config.GrpcEndpoint)
	assert.Equal(t, 30*time.Second, config.RequestTimeout)
	assert.Equal(t, 1*time.Second, config.RetryInterval)
	assert.Equal(t, 3, config.MaxRetries)
}

func TestClientClose(t *testing.T) {
	client := NewClientWithDefaults()

	// Should not panic
	err := client.Close()
	assert.NoError(t, err)
}

func TestTransactionCreation(t *testing.T) {
	client := NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte("test payload")

	// This would normally require a running Seata server
	// For now, we'll just test the client creation
	tx, err := client.StartTransaction(ctx, ModeSaga, payload)

	// Expect an error since no server is running
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestTransactionGetters(t *testing.T) {
	// Create a mock transaction
	client := NewClientWithDefaults()
	tx := &Transaction{
		client:   client,
		gid:      "test-gid",
		mode:     ModeSaga,
		payload:  []byte("test"),
		branches: []*Branch{},
	}

	assert.Equal(t, "test-gid", tx.GetGID())
	assert.Equal(t, ModeSaga, tx.GetMode())
	assert.Empty(t, tx.GetBranches())
}

func TestBranchCreation(t *testing.T) {
	branch := &Branch{
		BranchID: "branch-1",
		Action:   "http://example.com/action",
		Status:   BranchStatusPrepared,
	}

	assert.Equal(t, "branch-1", branch.BranchID)
	assert.Equal(t, "http://example.com/action", branch.Action)
	assert.Equal(t, BranchStatusPrepared, branch.Status)
}

func TestTransactionInfo(t *testing.T) {
	info := &TransactionInfo{
		GID:         "test-gid",
		Mode:        ModeSaga,
		Status:      StatusCommitted,
		Payload:     []byte("test"),
		Branches:    []Branch{},
		UpdatedUnix: 1234567890,
		CreatedUnix: 1234567890,
	}

	assert.Equal(t, "test-gid", info.GID)
	assert.Equal(t, ModeSaga, info.Mode)
	assert.Equal(t, StatusCommitted, info.Status)
	assert.Equal(t, []byte("test"), info.Payload)
	assert.Empty(t, info.Branches)
}

func TestHealthStatus(t *testing.T) {
	health := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	assert.Equal(t, "healthy", health.Status)
	assert.False(t, health.Timestamp.IsZero())
}

func TestSagaWorkflow(t *testing.T) {
	workflow := CreateSagaWorkflow([]SagaStep{
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

	assert.Len(t, workflow.Steps, 2)
	assert.Equal(t, "step1", workflow.Steps[0].BranchID)
	assert.Equal(t, "step2", workflow.Steps[1].BranchID)

	// Test validation
	err := workflow.Validate()
	assert.NoError(t, err)
}

func TestSagaWorkflowValidation(t *testing.T) {
	// Test empty workflow
	workflow := CreateSagaWorkflow([]SagaStep{})
	err := workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one step")

	// Test duplicate branch ID
	workflow = CreateSagaWorkflow([]SagaStep{
		{BranchID: "step1", Action: "http://example.com/step1"},
		{BranchID: "step1", Action: "http://example.com/step2"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate branch ID")

	// Test empty branch ID
	workflow = CreateSagaWorkflow([]SagaStep{
		{BranchID: "", Action: "http://example.com/step1"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch ID cannot be empty")

	// Test empty action
	workflow = CreateSagaWorkflow([]SagaStep{
		{BranchID: "step1", Action: ""},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action cannot be empty")
}

func TestTCCWorkflow(t *testing.T) {
	workflow := CreateTCCWorkflow([]TCCStep{
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

	assert.Len(t, workflow.Steps, 2)
	assert.Equal(t, "step1", workflow.Steps[0].BranchID)
	assert.Equal(t, "step2", workflow.Steps[1].BranchID)

	// Test validation
	err := workflow.Validate()
	assert.NoError(t, err)
}

func TestTCCWorkflowValidation(t *testing.T) {
	// Test empty workflow
	workflow := CreateTCCWorkflow([]TCCStep{})
	err := workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one step")

	// Test duplicate branch ID
	workflow = CreateTCCWorkflow([]TCCStep{
		{BranchID: "step1", Try: "http://example.com/step1/try", Confirm: "http://example.com/step1/confirm", Cancel: "http://example.com/step1/cancel"},
		{BranchID: "step1", Try: "http://example.com/step2/try", Confirm: "http://example.com/step2/confirm", Cancel: "http://example.com/step2/cancel"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate branch ID")

	// Test empty branch ID
	workflow = CreateTCCWorkflow([]TCCStep{
		{BranchID: "", Try: "http://example.com/step1/try", Confirm: "http://example.com/step1/confirm", Cancel: "http://example.com/step1/cancel"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch ID cannot be empty")

	// Test empty try
	workflow = CreateTCCWorkflow([]TCCStep{
		{BranchID: "step1", Try: "", Confirm: "http://example.com/step1/confirm", Cancel: "http://example.com/step1/cancel"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "try action cannot be empty")

	// Test empty confirm
	workflow = CreateTCCWorkflow([]TCCStep{
		{BranchID: "step1", Try: "http://example.com/step1/try", Confirm: "", Cancel: "http://example.com/step1/cancel"},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "confirm action cannot be empty")

	// Test empty cancel
	workflow = CreateTCCWorkflow([]TCCStep{
		{BranchID: "step1", Try: "http://example.com/step1/try", Confirm: "http://example.com/step1/confirm", Cancel: ""},
	})
	err = workflow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancel action cannot be empty")
}

func TestRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryInterval)
	assert.Equal(t, 2.0, config.BackoffFactor)
}

func TestCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	assert.Equal(t, 5, config.FailureThreshold)
	assert.Equal(t, 30*time.Second, config.RecoveryTimeout)
	assert.Equal(t, 3, config.HalfOpenMaxCalls)
}

func TestExecutionOptions(t *testing.T) {
	options := DefaultExecutionOptions()

	assert.Equal(t, 30*time.Second, options.Timeout)
	assert.NotNil(t, options.RetryConfig)
	assert.NotNil(t, options.CircuitBreaker)
	assert.True(t, options.ParallelBranches)
	assert.Equal(t, 10, options.MaxConcurrency)
}

func TestRetryManager(t *testing.T) {
	config := DefaultRetryConfig()
	retryManager := NewRetryManager(config)

	assert.NotNil(t, retryManager)
	assert.Equal(t, config, retryManager.config)
}

func TestCircuitBreaker(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	assert.NotNil(t, cb)
	assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	assert.Equal(t, 0, cb.failureCount)
}

func TestCircuitBreakerExecution(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.FailureThreshold = 2 // Lower threshold for testing
	cb := NewCircuitBreaker(config)

	// Test successful execution
	successCount := 0
	err := cb.Execute(func() error {
		successCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, successCount)
	assert.Equal(t, CircuitBreakerClosed, cb.GetState())

	// Test failure execution
	failureCount := 0
	err = cb.Execute(func() error {
		failureCount++
		return assert.AnError
	})

	assert.Error(t, err)
	assert.Equal(t, 1, failureCount)
	assert.Equal(t, CircuitBreakerClosed, cb.GetState()) // Still closed, need more failures

	// Test threshold exceeded
	err = cb.Execute(func() error {
		failureCount++
		return assert.AnError
	})

	assert.Error(t, err)
	assert.Equal(t, 2, failureCount)
	assert.Equal(t, CircuitBreakerOpen, cb.GetState())

	// Test reset
	cb.Reset()
	assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	assert.Equal(t, 0, cb.failureCount)
}
