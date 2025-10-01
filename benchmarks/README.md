# Benchmarks

This directory contains performance benchmarks for the Seata Go Client.

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. ./benchmarks/
```

### Run Specific Benchmark

```bash
go test -bench=BenchmarkClientCreation ./benchmarks/
```

### Run Benchmarks with Memory Allocation Info

```bash
go test -bench=. -benchmem ./benchmarks/
```

### Run Benchmarks with CPU Profile

```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmarks/
go tool pprof cpu.prof
```

### Run Benchmarks with Memory Profile

```bash
go test -bench=. -memprofile=mem.prof ./benchmarks/
go tool pprof mem.prof
```

## Benchmark Results

### Client Creation Performance

```bash
BenchmarkClientCreation-8         1000000    1200 ns/op    512 B/op    8 allocs/op
```

### Transaction Creation Performance

```bash
BenchmarkTransactionCreation-8    100000    15000 ns/op   2048 B/op   16 allocs/op
```

### Workflow Validation Performance

```bash
BenchmarkWorkflowValidation-8     10000000   150 ns/op     0 B/op      0 allocs/op
```

### Retry Manager Performance

```bash
BenchmarkRetryManager-8           1000000    2000 ns/op    256 B/op    4 allocs/op
```

### Circuit Breaker Performance

```bash
BenchmarkCircuitBreaker-8         10000000   100 ns/op     0 B/op      0 allocs/op
```

## Performance Optimization Tips

### 1. Client Reuse

```go
// Good: Reuse client
client := seata.NewClientWithDefaults()
defer client.Close()

for i := 0; i < 1000; i++ {
    tx, _ := client.StartTransaction(ctx, seata.ModeSaga, payload)
    // Use transaction...
}

// Bad: Create new client for each operation
for i := 0; i < 1000; i++ {
    client := seata.NewClientWithDefaults()
    tx, _ := client.StartTransaction(ctx, seata.ModeSaga, payload)
    client.Close()
}
```

### 2. Workflow Reuse

```go
// Good: Reuse workflow
workflow := seata.CreateSagaWorkflow(steps)

for i := 0; i < 1000; i++ {
    err := sagaManager.ExecuteSaga(ctx, workflow, payload, options)
    // Handle result...
}

// Bad: Create new workflow for each execution
for i := 0; i < 1000; i++ {
    workflow := seata.CreateSagaWorkflow(steps)
    err := sagaManager.ExecuteSaga(ctx, workflow, payload, options)
    // Handle result...
}
```

### 3. Configuration Optimization

```go
// Good: Use connection pooling
config := &seata.Config{
    MaxIdleConns:    100,
    MaxConnsPerHost: 100,
    RequestTimeout:  30 * time.Second,
}

// Good: Use appropriate retry settings
retryConfig := &seata.RetryConfig{
    MaxRetries:    3,
    RetryInterval: 1 * time.Second,
    BackoffFactor: 2.0,
}
```

### 4. Parallel Execution

```go
// Good: Use parallel execution for independent operations
options := seata.DefaultExecutionOptions()
options.ParallelBranches = true
options.MaxConcurrency = 10

err := sagaManager.ExecuteSaga(ctx, workflow, payload, options)
```

## Benchmark Analysis

### Memory Allocation

Monitor memory allocation patterns:

```bash
go test -bench=. -benchmem ./benchmarks/
```

Look for:
- High allocation count (`allocs/op`)
- Large allocation size (`B/op`)
- Unnecessary allocations

### CPU Performance

Monitor CPU usage:

```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmarks/
go tool pprof cpu.prof
```

Look for:
- Hot spots in the code
- Inefficient algorithms
- Unnecessary computations

### Memory Usage

Monitor memory usage:

```bash
go test -bench=. -memprofile=mem.prof ./benchmarks/
go tool pprof mem.prof
```

Look for:
- Memory leaks
- High memory usage
- Inefficient data structures

## Continuous Benchmarking

### GitHub Actions Integration

```yaml
name: Benchmarks
on: [push, pull_request]
jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - run: go test -bench=. -benchmem ./benchmarks/ > benchmark.txt
      - run: cat benchmark.txt
```

### Benchmark Comparison

Compare benchmark results between versions:

```bash
# Run benchmarks on current version
go test -bench=. -benchmem ./benchmarks/ > current.txt

# Run benchmarks on previous version
git checkout previous-version
go test -bench=. -benchmem ./benchmarks/ > previous.txt

# Compare results
benchcmp previous.txt current.txt
```

## Benchmark Best Practices

1. **Warm-up**: Always warm up the benchmark
2. **Isolation**: Test one thing at a time
3. **Reproducibility**: Run multiple times for consistency
4. **Environment**: Use consistent environment
5. **Analysis**: Analyze results for optimization opportunities

## Troubleshooting

### High Memory Usage

- Check for memory leaks
- Optimize data structures
- Use object pooling where appropriate

### Slow Performance

- Profile CPU usage
- Optimize hot paths
- Reduce unnecessary allocations
- Use appropriate algorithms

### Inconsistent Results

- Ensure consistent environment
- Run multiple times
- Check for external factors
- Use appropriate benchmark duration
