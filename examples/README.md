# Examples

All examples auto-start an ephemeral local mock HTTP server that returns 200 OK for branch actions. No fixed port is used.

Run format:

```bash
go run examples/*.go <example_name>
```

Core:
- basic
- saga
- tcc
- grpc
- comprehensive

Migrated (from seata-examples):
- mhttp_saga, mgrpc_saga, mgrpc_tcc
- mgrpc_headers, mgrpc_msg
- mhttp_headers, mhttp_msg
- mgrpc_saga_barrier, mgrpc_saga_other
- mhttp_workflow_saga, mhttp_workflow_tcc
- mhttp_xa, mhttp_gorm_barrier, mhttp_barrier_redis, mhttp_saga_mongo
- mgrpc_workflow_saga, mgrpc_workflow_tcc, mgrpc_workflow_mixed
- mhttp_saga_failure (new)
- mhttp_concurrent_saga (new)

Examples are idempotent and print final status (COMMITTED/ABORTED) when available.


