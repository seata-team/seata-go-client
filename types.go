package seata

import "time"

// HealthStatus represents the health status of the Seata server
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Transaction modes
const (
	ModeSaga = "saga"
	ModeTCC  = "tcc"
)

// Transaction statuses
const (
	StatusSubmitted = "SUBMITTED"
	StatusCommitted = "COMMITTED"
	StatusAborted   = "ABORTED"
)

// Branch statuses
const (
	BranchStatusPrepared = "PREPARED"
	BranchStatusSucceed  = "SUCCEED"
	BranchStatusFailed   = "FAILED"
)

// Error types
type SeataError struct {
	Code    string `json:"code"`
	Message string `json:"error"`
	Details string `json:"details"`
}

func (e *SeataError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeInvalidRequest      = "INVALID_REQUEST"
	ErrCodeTransactionNotFound = "TRANSACTION_NOT_FOUND"
	ErrCodeBranchNotFound      = "BRANCH_NOT_FOUND"
	ErrCodeServerError         = "SERVER_ERROR"
	ErrCodeTimeout             = "TIMEOUT"
	ErrCodeNetworkError        = "NETWORK_ERROR"
)

// Saga workflow helper types
type SagaStep struct {
	BranchID   string
	Action     string
	Compensate string
}

type SagaWorkflow struct {
	Steps []SagaStep
}

// TCC workflow helper types
type TCCStep struct {
	BranchID string
	Try      string
	Confirm  string
	Cancel   string
}

type TCCWorkflow struct {
	Steps []TCCStep
}

// Retry configuration
type RetryConfig struct {
	MaxRetries    int
	RetryInterval time.Duration
	BackoffFactor float64
}

// Default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	HalfOpenMaxCalls int
}

// Default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// Monitoring and metrics types
type Metrics struct {
	ActiveTransactions   int64   `json:"active_transactions"`
	BranchSuccessTotal   int64   `json:"branch_success_total"`
	BranchFailureTotal   int64   `json:"branch_failure_total"`
	BranchLatencySeconds float64 `json:"branch_latency_seconds"`
}

// Transaction statistics
type TransactionStats struct {
	TotalTransactions     int64 `json:"total_transactions"`
	CommittedTransactions int64 `json:"committed_transactions"`
	AbortedTransactions   int64 `json:"aborted_transactions"`
	ActiveTransactions    int64 `json:"active_transactions"`
}

// Branch execution result
type BranchResult struct {
	BranchID string
	Status   string
	Error    error
	Duration time.Duration
}

// Transaction execution options
type ExecutionOptions struct {
	Timeout          time.Duration
	RetryConfig      *RetryConfig
	CircuitBreaker   *CircuitBreakerConfig
	ParallelBranches bool
	MaxConcurrency   int
}

// Default execution options
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		Timeout:          30 * time.Second,
		RetryConfig:      DefaultRetryConfig(),
		CircuitBreaker:   DefaultCircuitBreakerConfig(),
		ParallelBranches: true,
		MaxConcurrency:   10,
	}
}
