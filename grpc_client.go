package seata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	seata_proto "github.com/seata-team/seata-go-client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient represents a gRPC client for Seata server
type GrpcClient struct {
	conn   *grpc.ClientConn
	client seata_proto.TransactionServiceClient
}

// NewGrpcClient creates a new gRPC client
func NewGrpcClient(endpoint string) *GrpcClient {
	client := &GrpcClient{}
	if err := client.Connect(endpoint); err != nil {
		// Log error but don't fail - connection will be established on first use
		fmt.Printf("Warning: Failed to connect to gRPC server: %v\n", err)
	}
	return client
}

// Connect establishes a connection to the gRPC server
func (gc *GrpcClient) Connect(endpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	target := endpoint
	if strings.HasPrefix(target, "grpc://") {
		target = strings.TrimPrefix(target, "grpc://")
	}
	if target == "" {
		return fmt.Errorf("invalid gRPC endpoint")
	}

	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	gc.conn = conn
	gc.client = seata_proto.NewTransactionServiceClient(conn)

	return nil
}

// Close closes the gRPC connection
func (gc *GrpcClient) Close() error {
	if gc.conn != nil {
		return gc.conn.Close()
	}
	return nil
}

// StartGlobal starts a global transaction via gRPC
func (gc *GrpcClient) StartGlobal(ctx context.Context, gid, mode string, payload []byte) (*seata_proto.StartGlobalResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.StartGlobalRequest{
		Gid:     gid,
		Mode:    mode,
		Payload: payload,
	}

	return gc.client.StartGlobal(ctx, req)
}

// Submit submits a transaction via gRPC
func (gc *GrpcClient) Submit(ctx context.Context, gid string) (*seata_proto.SubmitResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.SubmitRequest{
		Gid: gid,
	}

	return gc.client.Submit(ctx, req)
}

// Abort aborts a transaction via gRPC
func (gc *GrpcClient) Abort(ctx context.Context, gid string) (*seata_proto.AbortResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.AbortRequest{
		Gid: gid,
	}

	return gc.client.Abort(ctx, req)
}

// AddBranch adds a branch via gRPC
func (gc *GrpcClient) AddBranch(ctx context.Context, gid, branchID, action string) (*seata_proto.AddBranchResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.AddBranchRequest{
		Gid:      gid,
		BranchId: branchID,
		Action:   action,
	}

	return gc.client.AddBranch(ctx, req)
}

// BranchTry executes try phase via gRPC
func (gc *GrpcClient) BranchTry(ctx context.Context, gid, branchID, action string) (*seata_proto.BranchTryResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.BranchTryRequest{
		Gid:      gid,
		BranchId: branchID,
		Action:   action,
	}

	return gc.client.BranchTry(ctx, req)
}

// BranchSucceed marks branch as successful via gRPC
func (gc *GrpcClient) BranchSucceed(ctx context.Context, gid, branchID string) (*seata_proto.BranchStateResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.BranchStateRequest{
		Gid:      gid,
		BranchId: branchID,
	}

	return gc.client.BranchSucceed(ctx, req)
}

// BranchFail marks branch as failed via gRPC
func (gc *GrpcClient) BranchFail(ctx context.Context, gid, branchID string) (*seata_proto.BranchStateResponse, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.BranchStateRequest{
		Gid:      gid,
		BranchId: branchID,
	}

	return gc.client.BranchFail(ctx, req)
}

// Get retrieves a transaction via gRPC
func (gc *GrpcClient) Get(ctx context.Context, gid string) (*TransactionInfo, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.GetRequest{
		Gid: gid,
	}

	resp, err := gc.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var txInfo TransactionInfo
	if err := json.Unmarshal(resp.TxnJson, &txInfo); err != nil {
		return nil, fmt.Errorf("failed to parse transaction JSON: %w", err)
	}

	return &txInfo, nil
}

// List retrieves transactions via gRPC
func (gc *GrpcClient) List(ctx context.Context, limit, offset int, status string) ([]*TransactionInfo, error) {
	if gc.client == nil {
		return nil, fmt.Errorf("gRPC client not connected")
	}

	req := &seata_proto.ListRequest{
		Limit:  uint32(limit),
		Offset: uint32(offset),
		Status: status,
	}

	resp, err := gc.client.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse the JSON responses
	var transactions []*TransactionInfo
	for _, txnJson := range resp.TxnJson {
		var txInfo TransactionInfo
		if err := json.Unmarshal(txnJson, &txInfo); err != nil {
			return nil, fmt.Errorf("failed to parse transaction JSON: %w", err)
		}
		transactions = append(transactions, &txInfo)
	}

	return transactions, nil
}
