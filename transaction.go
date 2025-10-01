package seata

import (
	"context"
	"encoding/base64"
	"fmt"
)

// Transaction represents a global transaction
type Transaction struct {
	client   *Client
	gid      string
	mode     string
	payload  []byte
	branches []*Branch
}

// Branch represents a branch transaction
type Branch struct {
	BranchID string `json:"branch_id"`
	Action   string `json:"action"`
	Status   string `json:"status,omitempty"`
}

// TransactionInfo represents detailed transaction information
type TransactionInfo struct {
	GID         string   `json:"gid"`
	Mode        string   `json:"mode"`
	Status      string   `json:"status"`
	Payload     []byte   `json:"payload"`
	Branches    []Branch `json:"branches"`
	UpdatedUnix int64    `json:"updated_unix"`
	CreatedUnix int64    `json:"created_unix"`
}

// AddBranch adds a branch transaction to the global transaction
func (tx *Transaction) AddBranch(ctx context.Context, branchID, action string) error {
	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
		"action":    action,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/add")

	if err != nil {
		return fmt.Errorf("failed to add branch: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add branch: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	// Add branch to local list
	tx.branches = append(tx.branches, &Branch{
		BranchID: branchID,
		Action:   action,
	})

	return nil
}

// Submit submits the global transaction for execution
func (tx *Transaction) Submit(ctx context.Context) error {
	req := map[string]interface{}{
		"gid": tx.gid,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/submit")

	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to submit transaction: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// Abort aborts the global transaction
func (tx *Transaction) Abort(ctx context.Context) error {
	req := map[string]interface{}{
		"gid": tx.gid,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/abort")

	if err != nil {
		return fmt.Errorf("failed to abort transaction: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to abort transaction: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// GetGID returns the global transaction ID
func (tx *Transaction) GetGID() string {
	return tx.gid
}

// GetMode returns the transaction mode
func (tx *Transaction) GetMode() string {
	return tx.mode
}

// GetBranches returns the list of branches
func (tx *Transaction) GetBranches() []*Branch {
	return tx.branches
}

// TCC Transaction methods

// Try executes the try phase of a TCC branch
func (tx *Transaction) Try(ctx context.Context, branchID, action string, payload []byte) error {
	encodedPayload := base64.StdEncoding.EncodeToString(payload)

	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
		"action":    action,
		"payload":   encodedPayload,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/try")

	if err != nil {
		return fmt.Errorf("failed to execute try phase: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to execute try phase: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// Confirm executes the confirm phase of a TCC branch
func (tx *Transaction) Confirm(ctx context.Context, branchID string) error {
	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/confirm")

	if err != nil {
		return fmt.Errorf("failed to execute confirm phase: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to execute confirm phase: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// Cancel executes the cancel phase of a TCC branch
func (tx *Transaction) Cancel(ctx context.Context, branchID string) error {
	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/cancel")

	if err != nil {
		return fmt.Errorf("failed to execute cancel phase: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to execute cancel phase: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// BranchSucceed marks a branch as successful
func (tx *Transaction) BranchSucceed(ctx context.Context, branchID string) error {
	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/succeed")

	if err != nil {
		return fmt.Errorf("failed to mark branch as successful: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to mark branch as successful: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// BranchFail marks a branch as failed
func (tx *Transaction) BranchFail(ctx context.Context, branchID string) error {
	req := map[string]interface{}{
		"gid":       tx.gid,
		"branch_id": branchID,
	}

	resp, err := tx.client.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/branch/fail")

	if err != nil {
		return fmt.Errorf("failed to mark branch as failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to mark branch as failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// GetInfo retrieves the current transaction information
func (tx *Transaction) GetInfo(ctx context.Context) (*TransactionInfo, error) {
	return tx.client.GetTransaction(ctx, tx.gid)
}
