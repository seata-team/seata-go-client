package main

import (
	"context"
	"fmt"
	"log"

	"github.com/seata-team/seata-go-client"
)

// mhttp_gorm_barrier: simulate gorm barrier with OK endpoints
func mhttp_gorm_barrier() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_gorm_barrier"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "gb1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "gb2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_gorm_barrier finished")
}

// mhttp_barrier_redis: simulate redis barrier with OK endpoints
func mhttp_barrier_redis() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_barrier_redis"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "rb1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "rb2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_barrier_redis finished")
}

// mhttp_saga_mongo: simulate mongo saga with OK endpoints
func mhttp_saga_mongo() {
	baseURL, stop := startMockOKServer()
	defer stop()

	client := seata.NewClientWithDefaults()
	defer client.Close()

	ctx := context.Background()
	payload := []byte(`{"demo":"http_saga_mongo"}`)
	tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
	if err != nil {
		log.Fatalf("start failed: %v", err)
	}
	must(tx.AddBranch(ctx, "ms1", baseURL+"/ok"))
	must(tx.AddBranch(ctx, "ms2", baseURL+"/ok"))
	must(tx.Submit(ctx))
	waitStatus(ctx, client, tx.GetGID())
	fmt.Println("http_saga_mongo finished")
}
