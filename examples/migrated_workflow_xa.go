package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
)

// mhttp_workflow_saga: simplified HTTP workflow saga using OK endpoints
func mhttp_workflow_saga() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_workflow_saga"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "w1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "w2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_workflow_saga finished")
}

// mhttp_workflow_tcc: simplified HTTP workflow TCC using OK endpoints
func mhttp_workflow_tcc() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_workflow_tcc"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeTCC, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.Try(ctx, "tw1", baseURL+"/ok", payload))
	must(tx.Try(ctx, "tw2", baseURL+"/ok", payload))
	must(tx.Confirm(ctx, "tw1"))
	must(tx.Confirm(ctx, "tw2"))
	fmt.Println("http_workflow_tcc finished")
}

// mhttp_xa: placeholder XA style via saga (since XA not implemented client-side)
func mhttp_xa() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_xa"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "xa1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "xa2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_xa finished")
}

// mhttp_workflow_xa: simplified XA-like workflow using saga branches
func mhttp_workflow_xa() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_workflow_xa"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "wxa1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "wxa2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_workflow_xa finished")
}
