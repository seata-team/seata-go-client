package seata

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SagaManager provides high-level Saga pattern management
type SagaManager struct {
	client *Client
}

// NewSagaManager creates a new Saga manager
func NewSagaManager(client *Client) *SagaManager {
	return &SagaManager{
		client: client,
	}
}

// ExecuteSaga executes a complete Saga workflow
func (sm *SagaManager) ExecuteSaga(ctx context.Context, workflow *SagaWorkflow, payload []byte, options *ExecutionOptions) error {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Start global transaction
	tx, err := sm.client.StartTransaction(ctx, ModeSaga, payload)
	if err != nil {
		return fmt.Errorf("failed to start saga transaction: %w", err)
	}

	// Add all branches
	for _, step := range workflow.Steps {
		if err := tx.AddBranch(ctx, step.BranchID, step.Action); err != nil {
			// If adding branch fails, abort the transaction
			tx.Abort(ctx)
			return fmt.Errorf("failed to add branch %s: %w", step.BranchID, err)
		}
	}

	// Submit transaction for execution
	if err := tx.Submit(ctx); err != nil {
		return fmt.Errorf("failed to submit saga transaction: %w", err)
	}

	// Wait for completion and handle compensation if needed
	return sm.waitForCompletion(ctx, tx, workflow, options)
}

// ExecuteSagaWithCompensation executes a Saga with custom compensation logic
func (sm *SagaManager) ExecuteSagaWithCompensation(ctx context.Context, workflow *SagaWorkflow, payload []byte, compensationFunc func(ctx context.Context, failedStep *SagaStep) error, options *ExecutionOptions) error {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Start global transaction
	tx, err := sm.client.StartTransaction(ctx, ModeSaga, payload)
	if err != nil {
		return fmt.Errorf("failed to start saga transaction: %w", err)
	}

	// Add all branches
	for _, step := range workflow.Steps {
		if err := tx.AddBranch(ctx, step.BranchID, step.Action); err != nil {
			tx.Abort(ctx)
			return fmt.Errorf("failed to add branch %s: %w", step.BranchID, err)
		}
	}

	// Submit transaction
	if err := tx.Submit(ctx); err != nil {
		return fmt.Errorf("failed to submit saga transaction: %w", err)
	}

	// Monitor execution and handle compensation
	return sm.executeWithCompensation(ctx, tx, workflow, compensationFunc, options)
}

// waitForCompletion waits for transaction completion
func (sm *SagaManager) waitForCompletion(ctx context.Context, tx *Transaction, workflow *SagaWorkflow, options *ExecutionOptions) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(options.Timeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("saga execution timeout")
		case <-ticker.C:
			info, err := tx.GetInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get transaction info: %w", err)
			}

			switch info.Status {
			case StatusCommitted:
				return nil
			case StatusAborted:
				return fmt.Errorf("saga transaction aborted")
			case StatusSubmitted:
				// Still executing, continue waiting
				continue
			default:
				return fmt.Errorf("unknown transaction status: %s", info.Status)
			}
		}
	}
}

// executeWithCompensation executes saga with custom compensation
func (sm *SagaManager) executeWithCompensation(ctx context.Context, tx *Transaction, workflow *SagaWorkflow, compensationFunc func(ctx context.Context, failedStep *SagaStep) error, options *ExecutionOptions) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(options.Timeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("saga execution timeout")
		case <-ticker.C:
			info, err := tx.GetInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get transaction info: %w", err)
			}

			switch info.Status {
			case StatusCommitted:
				return nil
			case StatusAborted:
				// Find failed branches and execute compensation
				return sm.executeCompensation(ctx, workflow, info.Branches, compensationFunc)
			case StatusSubmitted:
				continue
			default:
				return fmt.Errorf("unknown transaction status: %s", info.Status)
			}
		}
	}
}

// executeCompensation executes compensation for failed steps
func (sm *SagaManager) executeCompensation(ctx context.Context, workflow *SagaWorkflow, branches []Branch, compensationFunc func(ctx context.Context, failedStep *SagaStep) error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Steps))

	// Find failed branches and execute compensation in reverse order
	for i := len(workflow.Steps) - 1; i >= 0; i-- {
		step := workflow.Steps[i]

		// Check if this branch failed
		var branchFailed bool
		for _, branch := range branches {
			if branch.BranchID == step.BranchID && branch.Status == BranchStatusFailed {
				branchFailed = true
				break
			}
		}

		if branchFailed {
			wg.Add(1)
			go func(step SagaStep) {
				defer wg.Done()
				if err := compensationFunc(ctx, &step); err != nil {
					errChan <- fmt.Errorf("compensation failed for branch %s: %w", step.BranchID, err)
				}
			}(step)
		}
	}

	wg.Wait()
	close(errChan)

	// Collect any compensation errors
	var compensationErrors []error
	for err := range errChan {
		compensationErrors = append(compensationErrors, err)
	}

	if len(compensationErrors) > 0 {
		return fmt.Errorf("compensation failed: %v", compensationErrors)
	}

	return nil
}

// CreateSagaWorkflow creates a new Saga workflow
func CreateSagaWorkflow(steps []SagaStep) *SagaWorkflow {
	return &SagaWorkflow{
		Steps: steps,
	}
}

// AddStep adds a step to the Saga workflow
func (sw *SagaWorkflow) AddStep(branchID, action, compensate string) {
	step := SagaStep{
		BranchID:   branchID,
		Action:     action,
		Compensate: compensate,
	}
	sw.Steps = append(sw.Steps, step)
}

// Validate validates the Saga workflow
func (sw *SagaWorkflow) Validate() error {
	if len(sw.Steps) == 0 {
		return fmt.Errorf("saga workflow must have at least one step")
	}

	seen := make(map[string]bool)
	for _, step := range sw.Steps {
		if step.BranchID == "" {
			return fmt.Errorf("branch ID cannot be empty")
		}
		if step.Action == "" {
			return fmt.Errorf("action cannot be empty")
		}
		if seen[step.BranchID] {
			return fmt.Errorf("duplicate branch ID: %s", step.BranchID)
		}
		seen[step.BranchID] = true
	}

	return nil
}
