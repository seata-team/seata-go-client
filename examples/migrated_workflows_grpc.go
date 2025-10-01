package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
)

// mgrpc_workflow_saga: gRPC control plane with HTTP OK branches
func mgrpc_workflow_saga() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_workflow_saga"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "ws1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "ws2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_workflow_saga finished")
}

// mgrpc_workflow_tcc: gRPC control plane with TCC try/confirm
func mgrpc_workflow_tcc() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_workflow_tcc"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.Try(ctx, "wt1", baseURL+"/ok", payload))
	must(tx.Try(ctx, "wt2", baseURL+"/ok", payload))
	must(tx.Confirm(ctx, "wt1"))
	must(tx.Confirm(ctx, "wt2"))
	fmt.Println("grpc_workflow_tcc finished")
}

// mgrpc_workflow_mixed: start saga then a TCC-like branch, demonstrating mixed patterns
func mgrpc_workflow_mixed() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_workflow_mixed"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	// saga branch
	must(tx.AddBranch(ctx, "mx1", baseURL+"/ok"))
	// tcc-like try/confirm using same transaction for demo purpose
	must(tx.Try(ctx, "mx2", baseURL+"/ok", payload))
	must(tx.Confirm(ctx, "mx2"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_workflow_mixed finished")
}
