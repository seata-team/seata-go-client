# API Reference

Complete API documentation for the Seata Go Client.

## Table of Contents

- [Client](#client)
- [Transaction](#transaction)
- [Saga Manager](#saga-manager)
- [TCC Manager](#tcc-manager)
- [Retry Manager](#retry-manager)
- [Circuit Breaker](#circuit-breaker)
- [Types](#types)
- [Constants](#constants)

## Client

### NewClient

```go
func NewClient(config *Config) *Client
```

Creates a new Seata client with the given configuration.

**Parameters:**
- `config *Config` - Client configuration

**Returns:**
- `*Client` - New client instance

**Example:**
```go
config := &seata.Config{
    HTTPEndpoint: "http://localhost:36789",
    RequestTimeout: 30 * time.Second,
}
client := seata.NewClient(config)
```

### NewClientWithDefaults

```go
func NewClientWithDefaults() *Client
```

Creates a new Seata client with default configuration.

**Returns:**
- `*Client` - New client instance

**Example:**
```go
client := seata.NewClientWithDefaults()
```

### StartTransaction

```go
func (c *Client) StartTransaction(ctx context.Context, mode string, payload []byte) (*Transaction, error)
```

Creates a new global transaction.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `mode string` - Transaction mode (`seata.ModeSaga` or `seata.ModeTCC`)
- `payload []byte` - Transaction payload

> Serialization: The server expects `payload` as a JSON integer array. The client automatically converts the provided `[]byte` to a `[]int` array during JSON serialization.

**Returns:**
- `*Transaction` - Transaction instance
- `error` - Error if transaction creation fails

**Example:**
```go
payload := []byte(`{"order_id": "12345", "amount": 100.00}`)
tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
if err != nil {
    log.Fatal(err)
}
```

### GetTransaction

```go
func (c *Client) GetTransaction(ctx context.Context, gid string) (*TransactionInfo, error)
```

Retrieves a specific transaction by its global ID.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `gid string` - Global transaction ID

**Returns:**
- `*TransactionInfo` - Transaction information
- `error` - Error if transaction not found

**Example:**
```go
info, err := client.GetTransaction(ctx, "transaction-id")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Transaction status: %s\n", info.Status)
```

### ListTransactions

```go
func (c *Client) ListTransactions(ctx context.Context, limit, offset int, status string) ([]*TransactionInfo, error)
```

Retrieves a list of transactions with optional filtering and pagination.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `limit int` - Number of transactions to return
- `offset int` - Number of transactions to skip
- `status string` - Filter by status (optional)

**Returns:**
- `[]*TransactionInfo` - List of transactions
- `error` - Error if listing fails

**Example:**
```go
transactions, err := client.ListTransactions(ctx, 10, 0, "COMMITTED")
if err != nil {
    log.Fatal(err)
}
```

### Health

```go
func (c *Client) Health(ctx context.Context) (*HealthStatus, error)
```

Checks the health of the Seata server.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout

**Returns:**
- `*HealthStatus` - Health status
- `error` - Error if health check fails

**Example:**
```go
health, err := client.Health(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Server status: %s\n", health.Status)
```

> Behavior: The server returns plain text `ok`. The client normalizes this to `HealthStatus{ Status: "healthy" }`.

### Metrics

```go
func (c *Client) Metrics(ctx context.Context) (string, error)
```

Retrieves Prometheus metrics from the server.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout

**Returns:**
- `string` - Prometheus metrics
- `error` - Error if metrics retrieval fails

**Example:**
```go
metrics, err := client.Metrics(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(metrics)
```

### Close

```go
func (c *Client) Close() error
```

Closes the client and releases resources.

**Returns:**
- `error` - Error if closing fails

**Example:**
```go
defer client.Close()
```

## Transaction

### AddBranch

```go
func (tx *Transaction) AddBranch(ctx context.Context, branchID, action string) error
```

Adds a branch transaction to the global transaction.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `branchID string` - Unique branch identifier
- `action string` - Action URL for the branch

**Returns:**
- `error` - Error if adding branch fails

**Example:**
```go
err := tx.AddBranch(ctx, "order-service", "http://order-service:8080/api/orders")
if err != nil {
    log.Fatal(err)
}
```

### Submit

```go
func (tx *Transaction) Submit(ctx context.Context) error
```

Submits the global transaction for execution.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout

**Returns:**
- `error` - Error if submission fails

**Example:**
```go
err := tx.Submit(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Abort

```go
func (tx *Transaction) Abort(ctx context.Context) error
```

Aborts the global transaction.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout

**Returns:**
- `error` - Error if abort fails

**Example:**
```go
err := tx.Abort(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Try (TCC)

```go
func (tx *Transaction) Try(ctx context.Context, branchID, action string, payload []byte) error
```

Executes the try phase of a TCC branch.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `branchID string` - Branch identifier
- `action string` - Try action URL
- `payload []byte` - Try payload

**Returns:**
- `error` - Error if try phase fails

**Example:**
```go
payload := []byte(`{"order_id": "12345"}`)
err := tx.Try(ctx, "order-service", "http://order-service:8080/api/orders/try", payload)
if err != nil {
    log.Fatal(err)
}
```

### Confirm (TCC)

```go
func (tx *Transaction) Confirm(ctx context.Context, branchID string) error
```

Executes the confirm phase of a TCC branch.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `branchID string` - Branch identifier

**Returns:**
- `error` - Error if confirm phase fails

**Example:**
```go
err := tx.Confirm(ctx, "order-service")
if err != nil {
    log.Fatal(err)
}
```

### Cancel (TCC)

```go
func (tx *Transaction) Cancel(ctx context.Context, branchID string) error
```

Executes the cancel phase of a TCC branch.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `branchID string` - Branch identifier

**Returns:**
- `error` - Error if cancel phase fails

**Example:**
```go
err := tx.Cancel(ctx, "order-service")
if err != nil {
    log.Fatal(err)
}
```

### GetGID

```go
func (tx *Transaction) GetGID() string
```

Returns the global transaction ID.

**Returns:**
- `string` - Global transaction ID

**Example:**
```go
gid := tx.GetGID()
fmt.Printf("Transaction ID: %s\n", gid)
```

### GetMode

```go
func (tx *Transaction) GetMode() string
```

Returns the transaction mode.

**Returns:**
- `string` - Transaction mode

**Example:**
```go
mode := tx.GetMode()
fmt.Printf("Transaction mode: %s\n", mode)
```

### GetBranches

```go
func (tx *Transaction) GetBranches() []*Branch
```

Returns the list of branches.

**Returns:**
- `[]*Branch` - List of branches

**Example:**
```go
branches := tx.GetBranches()
for _, branch := range branches {
    fmt.Printf("Branch: %s -> %s\n", branch.BranchID, branch.Action)
}
```

## Saga Manager

### NewSagaManager

```go
func NewSagaManager(client *Client) *SagaManager
```

Creates a new Saga manager.

**Parameters:**
- `client *Client` - Seata client

**Returns:**
- `*SagaManager` - Saga manager instance

**Example:**
```go
sagaManager := seata.NewSagaManager(client)
```

### ExecuteSaga

```go
func (sm *SagaManager) ExecuteSaga(ctx context.Context, workflow *SagaWorkflow, payload []byte, options *ExecutionOptions) error
```

Executes a complete Saga workflow.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `workflow *SagaWorkflow` - Saga workflow definition
- `payload []byte` - Workflow payload
- `options *ExecutionOptions` - Execution options

**Returns:**
- `error` - Error if execution fails

**Example:**
```go
workflow := seata.CreateSagaWorkflow(steps)
options := seata.DefaultExecutionOptions()
err := sagaManager.ExecuteSaga(ctx, workflow, payload, options)
```

### ExecuteSagaWithCompensation

```go
func (sm *SagaManager) ExecuteSagaWithCompensation(ctx context.Context, workflow *SagaWorkflow, payload []byte, compensationFunc func(ctx context.Context, failedStep *SagaStep) error, options *ExecutionOptions) error
```

Executes a Saga workflow with custom compensation logic.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `workflow *SagaWorkflow` - Saga workflow definition
- `payload []byte` - Workflow payload
- `compensationFunc func(ctx context.Context, failedStep *SagaStep) error` - Custom compensation function
- `options *ExecutionOptions` - Execution options

**Returns:**
- `error` - Error if execution fails

**Example:**
```go
compensationFunc := func(ctx context.Context, failedStep *seata.SagaStep) error {
    // Custom compensation logic
    return nil
}
err := sagaManager.ExecuteSagaWithCompensation(ctx, workflow, payload, compensationFunc, options)
```

## TCC Manager

### NewTCCManager

```go
func NewTCCManager(client *Client) *TCCManager
```

Creates a new TCC manager.

**Parameters:**
- `client *Client` - Seata client

**Returns:**
- `*TCCManager` - TCC manager instance

**Example:**
```go
tccManager := seata.NewTCCManager(client)
```

### ExecuteTCC

```go
func (tm *TCCManager) ExecuteTCC(ctx context.Context, workflow *TCCWorkflow, payload []byte, options *ExecutionOptions) error
```

Executes a complete TCC workflow.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `workflow *TCCWorkflow` - TCC workflow definition
- `payload []byte` - Workflow payload
- `options *ExecutionOptions` - Execution options

**Returns:**
- `error` - Error if execution fails

**Example:**
```go
workflow := seata.CreateTCCWorkflow(steps)
options := seata.DefaultExecutionOptions()
err := tccManager.ExecuteTCC(ctx, workflow, payload, options)
```

### ExecuteTCCWithBarrier

```go
func (tm *TCCManager) ExecuteTCCWithBarrier(ctx context.Context, workflow *TCCWorkflow, payload []byte, barrierID string, options *ExecutionOptions) error
```

Executes a TCC workflow with barrier pattern for idempotency.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `workflow *TCCWorkflow` - TCC workflow definition
- `payload []byte` - Workflow payload
- `barrierID string` - Barrier identifier for idempotency
- `options *ExecutionOptions` - Execution options

**Returns:**
- `error` - Error if execution fails

**Example:**
```go
barrierID := "barrier-12345"
err := tccManager.ExecuteTCCWithBarrier(ctx, workflow, payload, barrierID, options)
```

## Retry Manager

### NewRetryManager

```go
func NewRetryManager(config *RetryConfig) *RetryManager
```

Creates a new retry manager.

**Parameters:**
- `config *RetryConfig` - Retry configuration

**Returns:**
- `*RetryManager` - Retry manager instance

**Example:**
```go
retryConfig := seata.DefaultRetryConfig()
retryManager := seata.NewRetryManager(retryConfig)
```

### ExecuteWithRetry

```go
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation func() error) error
```

Executes an operation with retry logic.

**Parameters:**
- `ctx context.Context` - Context for cancellation and timeout
- `operation func() error` - Operation to execute

**Returns:**
- `error` - Error if all retries fail

**Example:**
```go
err := retryManager.ExecuteWithRetry(ctx, func() error {
    return someOperation()
})
```

## Circuit Breaker

### NewCircuitBreaker

```go
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker
```

Creates a new circuit breaker.

**Parameters:**
- `config *CircuitBreakerConfig` - Circuit breaker configuration

**Returns:**
- `*CircuitBreaker` - Circuit breaker instance

**Example:**
```go
config := seata.DefaultCircuitBreakerConfig()
circuitBreaker := seata.NewCircuitBreaker(config)
```

### Execute

```go
func (cb *CircuitBreaker) Execute(operation func() error) error
```

Executes an operation through the circuit breaker.

**Parameters:**
- `operation func() error` - Operation to execute

**Returns:**
- `error` - Error if operation fails or circuit is open

**Example:**
```go
err := circuitBreaker.Execute(func() error {
    return someOperation()
})
```

## Types

### Config

```go
type Config struct {
    HTTPEndpoint    string
    GrpcEndpoint    string
    RequestTimeout  time.Duration
    RetryInterval   time.Duration
    MaxRetries      int
    MaxIdleConns    int
    MaxConnsPerHost int
    AuthToken       string
}
```

Client configuration.

### TransactionInfo

```go
type TransactionInfo struct {
    GID         string   `json:"gid"`
    Mode        string   `json:"mode"`
    Status      string   `json:"status"`
    Payload     []byte   `json:"payload"`
    Branches    []Branch `json:"branches"`
    UpdatedUnix int64    `json:"updated_unix"`
    CreatedUnix int64    `json:"created_unix"`
}
```

Transaction information.

### SagaStep

```go
type SagaStep struct {
    BranchID   string
    Action     string
    Compensate string
}
```

Saga workflow step.

### TCCStep

```go
type TCCStep struct {
    BranchID string
    Try      string
    Confirm  string
    Cancel   string
}
```

TCC workflow step.

## Constants

### Transaction Modes

```go
const (
    ModeSaga = "saga"
    ModeTCC  = "tcc"
)
```

### Transaction Statuses

```go
const (
    StatusSubmitted = "SUBMITTED"
    StatusCommitted = "COMMITTED"
    StatusAborted   = "ABORTED"
)
```

### Branch Statuses

```go
const (
    BranchStatusPrepared = "PREPARED"
    BranchStatusSucceed  = "SUCCEED"
    BranchStatusFailed   = "FAILED"
)
```

### Error Codes

```go
const (
    ErrCodeInvalidRequest     = "INVALID_REQUEST"
    ErrCodeTransactionNotFound = "TRANSACTION_NOT_FOUND"
    ErrCodeBranchNotFound     = "BRANCH_NOT_FOUND"
    ErrCodeServerError        = "SERVER_ERROR"
    ErrCodeTimeout            = "TIMEOUT"
    ErrCodeNetworkError       = "NETWORK_ERROR"
)
```
