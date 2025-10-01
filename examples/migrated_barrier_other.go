package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
)

// migratedGrpcSagaBarrier: simplified barrier-like saga using OK endpoints
func migratedGrpcSagaBarrier() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_saga_barrier"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}

	// emulate barrier branches
	must(tx.AddBranch(ctx, "b1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "b2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_saga_barrier finished")
}

// migratedGrpcSagaOther: a minimal variant with different branch naming
func migratedGrpcSagaOther() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_saga_other"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}

	must(tx.AddBranch(ctx, "o1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "o2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_saga_other finished")
}
