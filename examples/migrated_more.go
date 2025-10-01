package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
	"google.golang.org/grpc/metadata"
)

// migratedGrpcHeaders demonstrates sending headers via HTTP and gRPC
func migratedGrpcHeaders() {
	baseURL, stop := startMockOKServer()
	defer stop()

	cfg := seata.DefaultConfig()
	cfg.GrpcEndpoint = "localhost:36790"
	client := seata.NewClient(cfg)
	defer client.Close()

	// add a simple header on HTTP path (health)
	ctx := context.Background()
	_, _ = client.Health(ctx) // warmup

	// gRPC metadata demo: attach dummy headers when starting
	md := metadata.New(map[string]string{"x-demo": "headers"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	payload := []byte(`{"demo":"grpc_headers"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "h1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "h2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(context.Background(), client, tx.GetGID())
}

// migratedGrpcMsg demonstrates a minimal message-like flow using Saga with payload only
func migratedGrpcMsg() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	// treat payload as message body
	payload := []byte(`{"msg":"hello"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "m1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "m2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("grpc_msg-like flow finished")
}

// migratedHttpHeaders mirrors the header passing pattern on HTTP branches
func migratedHttpHeaders() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_headers"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	// Headers are typically sent by the branch service; here we just hit OK endpoints
	must(tx.AddBranch(ctx, "h1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "h2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
}

// migratedHttpMsg simulates a message-style HTTP saga using simple OK branches
func migratedHttpMsg() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"msg":"hello-http"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "m1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "m2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
}
