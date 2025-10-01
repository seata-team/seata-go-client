package seata

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryManager handles retry logic for operations
type RetryManager struct {
	config *RetryConfig
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config *RetryConfig) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryManager{
		config: config,
	}
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute operation
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == rm.config.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := rm.calculateBackoff(attempt)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", rm.config.MaxRetries, lastErr)
}

// ExecuteWithRetryAndValidation executes a function with retry logic and validation
func (rm *RetryManager) ExecuteWithRetryAndValidation(ctx context.Context, operation func() error, validator func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute operation
		err := operation()
		if err != nil {
			lastErr = err

			// Don't retry on the last attempt
			if attempt == rm.config.MaxRetries {
				break
			}

			// Calculate backoff delay
			delay := rm.calculateBackoff(attempt)

			// Wait with context cancellation support
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
			continue
		}

		// Operation succeeded, validate result
		if validator != nil {
			if err := validator(); err != nil {
				lastErr = err

				// Don't retry on the last attempt
				if attempt == rm.config.MaxRetries {
					break
				}

				// Calculate backoff delay
				delay := rm.calculateBackoff(attempt)

				// Wait with context cancellation support
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					// Continue to next attempt
				}
				continue
			}
		}

		// Both operation and validation succeeded
		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %w", rm.config.MaxRetries, lastErr)
}

// calculateBackoff calculates the backoff delay for the given attempt
func (rm *RetryManager) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	baseDelay := float64(rm.config.RetryInterval)
	exponentialDelay := baseDelay * math.Pow(rm.config.BackoffFactor, float64(attempt))

	// Add jitter to prevent thundering herd
	jitter := time.Duration(float64(exponentialDelay) * 0.1 * (0.5 - math.Mod(float64(time.Now().UnixNano()), 1.0)))

	return time.Duration(exponentialDelay) + jitter
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a retryable error type
	// This is a placeholder implementation
	// In a real implementation, you would check the error type
	if fmt.Errorf("").Error() != "" { // Placeholder for actual implementation
		return true
	}

	return false
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation struct {
	Operation   func() error
	Validator   func() error
	IsRetryable func(error) bool
}

// Execute executes the retryable operation
func (ro *RetryableOperation) Execute(ctx context.Context, retryManager *RetryManager) error {
	return retryManager.ExecuteWithRetryAndValidation(ctx, ro.Operation, ro.Validator)
}

// CreateRetryableOperation creates a new retryable operation
func CreateRetryableOperation(operation func() error, validator func() error, isRetryable func(error) bool) *RetryableOperation {
	return &RetryableOperation{
		Operation:   operation,
		Validator:   validator,
		IsRetryable: isRetryable,
	}
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	failureCount    int
	lastFailureTime time.Time
	state           CircuitBreakerState
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	// Check circuit breaker state
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	// Execute operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	// Operation succeeded
	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and updates circuit breaker state
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.config.FailureThreshold {
		cb.state = CircuitBreakerOpen
	}
}

// recordSuccess records a success and resets circuit breaker state
func (cb *CircuitBreaker) recordSuccess() {
	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}
