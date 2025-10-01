package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seata-team/seata-go-client"
)

// migratedHttpSaga emulates seata-examples/examples/http_saga.go using our client
func migratedHttpSaga() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_saga"}`)

	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}

	must(tx.AddBranch(ctx, "step1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "step2", baseURL+"/ok"))

	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
}

// migratedGrpcSaga emulates grpc_saga.go at high level
func migratedGrpcSaga() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_saga"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "gstep1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "gstep2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
}

// migratedGrpcTcc emulates grpc_tcc.go high-level flow via Try/Confirm
func migratedGrpcTcc() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_tcc"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}

	// Try phases
	must(tx.Try(ctx, "tstep1", baseURL+"/ok", payload))
	must(tx.Try(ctx, "tstep2", baseURL+"/ok", payload))
	// Confirm phases
	must(tx.Confirm(ctx, "tstep1"))
	must(tx.Confirm(ctx, "tstep2"))
	fmt.Println("TCC flow finished")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func waitStatus(ctx context.Context, client *seata.Client, gid string) {
	for i := 0; i < 20; i++ {
		info, err := client.GetTransaction(ctx, gid)
		if err == nil {
			fmt.Println("status:", info.Status)
			if info.Status == seata.StatusCommitted || info.Status == seata.StatusAborted {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// mhttp_saga_failure: one branch succeeds, another fails -> expect ABORTED
func mhttp_saga_failure() {
    baseURL, stop := startMockOKServer()
    defer stop()

    client := seata.NewClientWithDefaults()
    defer client.Close()

    ctx := context.Background()
    payload := []byte(`{"demo":"http_saga_failure"}`)
    tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
    if err != nil {
        log.Fatalf("start failed: %v", err)
    }
    must(tx.AddBranch(ctx, "ok1", baseURL+"/ok"))
    must(tx.AddBranch(ctx, "bad", baseURL+"/fail"))
    _ = tx.Submit(ctx)
    for i := 0; i < 20; i++ {
        info, err := client.GetTransaction(ctx, tx.GetGID())
        if err == nil && (info.Status == seata.StatusAborted || info.Status == seata.StatusCommitted) {
            fmt.Println("status:", info.Status)
            break
        }
        time.Sleep(300 * time.Millisecond)
    }
}

// mhttp_concurrent_saga: simulate multiple branches and quick submit
func mhttp_concurrent_saga() {
    baseURL, stop := startMockOKServer()
    defer stop()

    client := seata.NewClientWithDefaults()
    defer client.Close()

    ctx := context.Background()
    payload := []byte(`{"demo":"http_concurrent_saga"}`)
    tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
    if err != nil {
        log.Fatalf("start failed: %v", err)
    }
    // Add several branches sequentially; concurrency is server-side
    for i := 0; i < 4; i++ {
        must(tx.AddBranch(ctx, fmt.Sprintf("c%d", i+1), baseURL+"/ok"))
    }
    must(tx.Submit(ctx))
    waitStatus(ctx, client, tx.GetGID())
}
