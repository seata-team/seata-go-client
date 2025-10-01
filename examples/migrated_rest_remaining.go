package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
)

// mhttp_tcc: HTTP TCC demo using OK endpoints
func mhttp_tcc() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_tcc"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.Try(ctx, "ht1", baseURL+"/ok", payload))
	must(tx.Try(ctx, "ht2", baseURL+"/ok", payload))
	must(tx.Confirm(ctx, "ht1"))
	must(tx.Confirm(ctx, "ht2"))
	fmt.Println("http_tcc finished")
}

// mhttp_tcc_barrier: simulate tcc with barrier concept (idempotency mocked)
func mhttp_tcc_barrier() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_tcc_barrier"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.Try(ctx, "hb1", baseURL+"/ok", payload))
	must(tx.Try(ctx, "hb2", baseURL+"/ok", payload))
	must(tx.Confirm(ctx, "hb1"))
	must(tx.Confirm(ctx, "hb2"))
	fmt.Println("http_tcc_barrier finished")
}

// mhttp_saga_barrier: saga with barrier naming (mocked)
func mhttp_saga_barrier() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_saga_barrier"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "sb1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "sb2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_saga_barrier finished")
}

// mhttp_saga_mutidb: saga hitting multiple DBs (simulated)
func mhttp_saga_mutidb() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_saga_mutidb"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "db1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "db2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_saga_mutidb finished")
}

// mhttp_xa_gorm: XA via gorm (simulated by saga branches)
func mhttp_xa_gorm() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_xa_gorm"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "xg1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "xg2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_xa_gorm finished")
}

// mgrpc_xa: gRPC control with XA-like flow (simulated)
func mgrpc_xa() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"grpc_xa"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "gx1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "gx2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_xa finished")
}

// mhttp_more: miscellaneous HTTP example (simulated)
func mhttp_more() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_more"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "m1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "m2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_more finished")
}

// mhttp_saga_redis: simulate redis barrier style saga with OK endpoints
func mhttp_saga_redis() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_saga_redis"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	// emulate two redis-protected branches
	must(tx.AddBranch(ctx, "rs1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "rs2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_saga_redis finished")
}
