package seata

import (
	"context"
	"fmt"
	"sync"
)

// TCCManager provides high-level TCC pattern management
type TCCManager struct {
	client *Client
}

// NewTCCManager creates a new TCC manager
func NewTCCManager(client *Client) *TCCManager {
	return &TCCManager{
		client: client,
	}
}

// ExecuteTCC executes a complete TCC workflow
func (tm *TCCManager) ExecuteTCC(ctx context.Context, workflow *TCCWorkflow, payload []byte, options *ExecutionOptions) error {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Start global transaction
	tx, err := tm.client.StartTransaction(ctx, ModeTCC, payload)
	if err != nil {
		return fmt.Errorf("failed to start TCC transaction: %w", err)
	}

	// Execute try phase for all branches
	if err := tm.executeTryPhase(ctx, tx, workflow, payload, options); err != nil {
		// Try phase failed, execute cancel phase for all branches
		tm.executeCancelPhase(ctx, tx, workflow)
		return fmt.Errorf("TCC try phase failed: %w", err)
	}

	// Try phase succeeded, execute confirm phase
	if err := tm.executeConfirmPhase(ctx, tx, workflow, options); err != nil {
		// Confirm phase failed, execute cancel phase
		tm.executeCancelPhase(ctx, tx, workflow)
		return fmt.Errorf("TCC confirm phase failed: %w", err)
	}

	return nil
}

// ExecuteTCCWithBarrier executes TCC with barrier pattern for idempotency
func (tm *TCCManager) ExecuteTCCWithBarrier(ctx context.Context, workflow *TCCWorkflow, payload []byte, barrierID string, options *ExecutionOptions) error {
	if options == nil {
		options = DefaultExecutionOptions()
	}

	// Start global transaction
	tx, err := tm.client.StartTransaction(ctx, ModeTCC, payload)
	if err != nil {
		return fmt.Errorf("failed to start TCC transaction: %w", err)
	}

	// Execute try phase with barrier
	if err := tm.executeTryPhaseWithBarrier(ctx, tx, workflow, payload, barrierID, options); err != nil {
		tm.executeCancelPhase(ctx, tx, workflow)
		return fmt.Errorf("TCC try phase with barrier failed: %w", err)
	}

	// Execute confirm phase with barrier
	if err := tm.executeConfirmPhaseWithBarrier(ctx, tx, workflow, barrierID, options); err != nil {
		tm.executeCancelPhase(ctx, tx, workflow)
		return fmt.Errorf("TCC confirm phase with barrier failed: %w", err)
	}

	return nil
}

// executeTryPhase executes the try phase for all branches
func (tm *TCCManager) executeTryPhase(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, payload []byte, options *ExecutionOptions) error {
	if options.ParallelBranches {
		return tm.executeTryPhaseParallel(ctx, tx, workflow, payload, options)
	}
	return tm.executeTryPhaseSequential(ctx, tx, workflow, payload, options)
}

// executeTryPhaseParallel executes try phase in parallel
func (tm *TCCManager) executeTryPhaseParallel(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, payload []byte, options *ExecutionOptions) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Steps))
	semaphore := make(chan struct{}, options.MaxConcurrency)

	for _, step := range workflow.Steps {
		wg.Add(1)
		go func(step TCCStep) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := tx.Try(ctx, step.BranchID, step.Try, payload); err != nil {
				errChan <- fmt.Errorf("try phase failed for branch %s: %w", step.BranchID, err)
			}
		}(step)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeTryPhaseSequential executes try phase sequentially
func (tm *TCCManager) executeTryPhaseSequential(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, payload []byte, options *ExecutionOptions) error {
	for _, step := range workflow.Steps {
		if err := tx.Try(ctx, step.BranchID, step.Try, payload); err != nil {
			return fmt.Errorf("try phase failed for branch %s: %w", step.BranchID, err)
		}
	}
	return nil
}

// executeTryPhaseWithBarrier executes try phase with barrier pattern
func (tm *TCCManager) executeTryPhaseWithBarrier(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, payload []byte, barrierID string, options *ExecutionOptions) error {
	// Add barrier ID to payload
	barrierPayload := append(payload, []byte(barrierID)...)

	if options.ParallelBranches {
		return tm.executeTryPhaseParallel(ctx, tx, workflow, barrierPayload, options)
	}
	return tm.executeTryPhaseSequential(ctx, tx, workflow, barrierPayload, options)
}

// executeConfirmPhase executes the confirm phase for all branches
func (tm *TCCManager) executeConfirmPhase(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, options *ExecutionOptions) error {
	if options.ParallelBranches {
		return tm.executeConfirmPhaseParallel(ctx, tx, workflow, options)
	}
	return tm.executeConfirmPhaseSequential(ctx, tx, workflow, options)
}

// executeConfirmPhaseParallel executes confirm phase in parallel
func (tm *TCCManager) executeConfirmPhaseParallel(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, options *ExecutionOptions) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Steps))
	semaphore := make(chan struct{}, options.MaxConcurrency)

	for _, step := range workflow.Steps {
		wg.Add(1)
		go func(step TCCStep) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := tx.Confirm(ctx, step.BranchID); err != nil {
				errChan <- fmt.Errorf("confirm phase failed for branch %s: %w", step.BranchID, err)
			}
		}(step)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeConfirmPhaseSequential executes confirm phase sequentially
func (tm *TCCManager) executeConfirmPhaseSequential(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, options *ExecutionOptions) error {
	for _, step := range workflow.Steps {
		if err := tx.Confirm(ctx, step.BranchID); err != nil {
			return fmt.Errorf("confirm phase failed for branch %s: %w", step.BranchID, err)
		}
	}
	return nil
}

// executeConfirmPhaseWithBarrier executes confirm phase with barrier pattern
func (tm *TCCManager) executeConfirmPhaseWithBarrier(ctx context.Context, tx *Transaction, workflow *TCCWorkflow, barrierID string, options *ExecutionOptions) error {
	// For barrier pattern, we need to ensure idempotency
	// This is typically handled by the service implementation
	return tm.executeConfirmPhase(ctx, tx, workflow, options)
}

// executeCancelPhase executes the cancel phase for all branches
func (tm *TCCManager) executeCancelPhase(ctx context.Context, tx *Transaction, workflow *TCCWorkflow) {
	var wg sync.WaitGroup

	for _, step := range workflow.Steps {
		wg.Add(1)
		go func(step TCCStep) {
			defer wg.Done()
			// Execute cancel phase (ignore errors for cleanup)
			tx.Cancel(ctx, step.BranchID)
		}(step)
	}

	wg.Wait()
}

// CreateTCCWorkflow creates a new TCC workflow
func CreateTCCWorkflow(steps []TCCStep) *TCCWorkflow {
	return &TCCWorkflow{
		Steps: steps,
	}
}

// AddStep adds a step to the TCC workflow
func (tw *TCCWorkflow) AddStep(branchID, try, confirm, cancel string) {
	step := TCCStep{
		BranchID: branchID,
		Try:      try,
		Confirm:  confirm,
		Cancel:   cancel,
	}
	tw.Steps = append(tw.Steps, step)
}

// Validate validates the TCC workflow
func (tw *TCCWorkflow) Validate() error {
	if len(tw.Steps) == 0 {
		return fmt.Errorf("TCC workflow must have at least one step")
	}

	seen := make(map[string]bool)
	for _, step := range tw.Steps {
		if step.BranchID == "" {
			return fmt.Errorf("branch ID cannot be empty")
		}
		if step.Try == "" {
			return fmt.Errorf("try action cannot be empty")
		}
		if step.Confirm == "" {
			return fmt.Errorf("confirm action cannot be empty")
		}
		if step.Cancel == "" {
			return fmt.Errorf("cancel action cannot be empty")
		}
		if seen[step.BranchID] {
			return fmt.Errorf("duplicate branch ID: %s", step.BranchID)
		}
		seen[step.BranchID] = true
	}

	return nil
}
