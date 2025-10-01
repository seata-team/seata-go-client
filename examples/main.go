package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <example>")
		fmt.Println("Available examples:")
		fmt.Println("  basic        - Basic client usage")
		fmt.Println("  saga         - Saga pattern example")
		fmt.Println("  tcc          - TCC pattern example")
		fmt.Println("  grpc         - gRPC client example")
		fmt.Println("  comprehensive - Comprehensive example with all features")
		fmt.Println("  mhttp_saga   - Migrated: HTTP Saga flow")
		fmt.Println("  mgrpc_saga   - Migrated: gRPC Saga flow")
		fmt.Println("  mgrpc_tcc    - Migrated: gRPC TCC flow")
		fmt.Println("  mgrpc_headers- Migrated: gRPC headers flow")
		fmt.Println("  mgrpc_msg    - Migrated: gRPC msg-like flow")
		fmt.Println("  mhttp_headers- Migrated: HTTP headers flow")
		fmt.Println("  mhttp_msg    - Migrated: HTTP msg-like flow")
		fmt.Println("  mgrpc_saga_barrier - Migrated: gRPC saga barrier flow")
		fmt.Println("  mgrpc_saga_other   - Migrated: gRPC saga other flow")
		fmt.Println("  mhttp_workflow_saga- Migrated: HTTP workflow saga")
		fmt.Println("  mhttp_workflow_tcc - Migrated: HTTP workflow TCC")
		fmt.Println("  mhttp_workflow_xa  - Migrated: HTTP workflow XA (sim)")
		fmt.Println("  mhttp_xa           - Migrated: HTTP XA (simulated)")
		fmt.Println("  mhttp_gorm_barrier - Migrated: HTTP GORM barrier (sim)")
		fmt.Println("  mhttp_barrier_redis- Migrated: HTTP Redis barrier (sim)")
		fmt.Println("  mhttp_saga_mongo   - Migrated: HTTP Saga Mongo (sim)")
		fmt.Println("  mhttp_saga_redis   - Migrated: HTTP Saga Redis (sim)")
		fmt.Println("  mgrpc_workflow_saga- Migrated: gRPC workflow saga")
		fmt.Println("  mgrpc_workflow_tcc - Migrated: gRPC workflow TCC")
		fmt.Println("  mgrpc_workflow_mixed - Migrated: gRPC workflow mixed")
		fmt.Println("  mhttp_tcc           - Migrated: HTTP TCC")
		fmt.Println("  mhttp_tcc_barrier   - Migrated: HTTP TCC barrier (sim)")
		fmt.Println("  mhttp_saga_barrier  - Migrated: HTTP Saga barrier (sim)")
		fmt.Println("  mhttp_saga_mutidb   - Migrated: HTTP Saga multi-DB (sim)")
		fmt.Println("  mhttp_xa_gorm       - Migrated: HTTP XA GORM (sim)")
		fmt.Println("  mgrpc_xa            - Migrated: gRPC XA (sim)")
		fmt.Println("  mhttp_more          - Migrated: HTTP more (sim)")
		fmt.Println("  mhttp_saga_failure  - New: HTTP Saga failure (sim)")
		fmt.Println("  mhttp_concurrent_saga - New: HTTP concurrent saga (sim)")
		return
	}

	example := os.Args[1]

	switch example {
	case "basic":
		basicExample()
	case "saga":
		sagaExample()
	case "tcc":
		tccExample()
	case "grpc":
		grpcExample()
	case "comprehensive":
		comprehensiveExample()
	case "mhttp_saga":
		migratedHttpSaga()
	case "mgrpc_saga":
		migratedGrpcSaga()
	case "mgrpc_tcc":
		migratedGrpcTcc()
	case "mgrpc_headers":
		migratedGrpcHeaders()
	case "mgrpc_msg":
		migratedGrpcMsg()
	case "mhttp_headers":
		migratedHttpHeaders()
	case "mhttp_msg":
		migratedHttpMsg()
	case "mgrpc_saga_barrier":
		migratedGrpcSagaBarrier()
	case "mgrpc_saga_other":
		migratedGrpcSagaOther()
	case "mhttp_workflow_saga":
		mhttp_workflow_saga()
	case "mhttp_workflow_tcc":
		mhttp_workflow_tcc()
	case "mhttp_workflow_xa":
		mhttp_workflow_xa()
	case "mhttp_xa":
		mhttp_xa()
	case "mhttp_gorm_barrier":
		mhttp_gorm_barrier()
	case "mhttp_barrier_redis":
		mhttp_barrier_redis()
	case "mhttp_saga_mongo":
		mhttp_saga_mongo()
	case "mhttp_saga_redis":
		mhttp_saga_redis()
	case "mgrpc_workflow_saga":
		mgrpc_workflow_saga()
	case "mgrpc_workflow_tcc":
		mgrpc_workflow_tcc()
	case "mgrpc_workflow_mixed":
		mgrpc_workflow_mixed()
	case "mhttp_tcc":
		mhttp_tcc()
	case "mhttp_tcc_barrier":
		mhttp_tcc_barrier()
	case "mhttp_saga_barrier":
		mhttp_saga_barrier()
	case "mhttp_saga_mutidb":
		mhttp_saga_mutidb()
	case "mhttp_xa_gorm":
		mhttp_xa_gorm()
	case "mgrpc_xa":
		mgrpc_xa()
	case "mhttp_more":
		mhttp_more()
	case "mhttp_saga_failure":
		mhttp_saga_failure()
	case "mhttp_concurrent_saga":
		mhttp_concurrent_saga()
	default:
		log.Fatalf("Unknown example: %s", example)
	}
}
