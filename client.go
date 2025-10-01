package seata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client represents a Seata client for distributed transaction management
type Client struct {
	httpClient *resty.Client
	grpcClient *GrpcClient
	config     *Config
	discovery  *EtcdDiscovery
	// lb state
	httpAddrs []string
	grpcAddrs []string
	lbIndex   int
	lbStop    chan struct{}
}

// bytesToIntArray converts a byte slice to an int slice for JSON serialization
func bytesToIntArray(source []byte) []int {
	result := make([]int, len(source))
	for index, value := range source {
		result[index] = int(value)
	}
	return result
}

// Config holds the client configuration
type Config struct {
	// Server configuration
	HTTPEndpoint string
	GrpcEndpoint string

	// Timeout settings
	RequestTimeout time.Duration
	RetryInterval  time.Duration
	MaxRetries     int

	// Connection settings
	MaxIdleConns    int
	MaxConnsPerHost int

	// Authentication (for future use)
	AuthToken string

	// Optional service discovery using etcd
	Discovery *DiscoveryConfig
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		HTTPEndpoint:    "http://localhost:36789",
		GrpcEndpoint:    "localhost:36790",
		RequestTimeout:  30 * time.Second,
		RetryInterval:   1 * time.Second,
		MaxRetries:      3,
		MaxIdleConns:    100,
		MaxConnsPerHost: 100,
		Discovery:       nil,
	}
}

// DiscoveryConfig holds etcd service discovery settings
type DiscoveryConfig struct {
	EtcdEndpoints []string
	Namespace     string // e.g. "/seata"
}

// NewClient creates a new Seata client with the given configuration
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client
	httpClient := resty.New()
	httpClient.SetBaseURL(config.HTTPEndpoint)
	httpClient.SetTimeout(config.RequestTimeout)
	httpClient.SetRetryCount(config.MaxRetries)
	httpClient.SetRetryWaitTime(config.RetryInterval)
	httpClient.SetRetryMaxWaitTime(config.RetryInterval * 3)

	// Set connection pool settings
	httpClient.GetClient().Transport = &http.Transport{
		MaxIdleConns:       config.MaxIdleConns,
		MaxConnsPerHost:    config.MaxConnsPerHost,
		IdleConnTimeout:    90 * time.Second,
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	// Create gRPC client
	grpcClient := NewGrpcClient(config.GrpcEndpoint)

	c := &Client{
		httpClient: httpClient,
		grpcClient: grpcClient,
		config:     config,
		lbStop:     make(chan struct{}),
	}

	// Start discovery if configured
	if config.Discovery != nil && len(config.Discovery.EtcdEndpoints) > 0 {
		d := NewEtcdDiscovery(config.Discovery.EtcdEndpoints, config.Discovery.Namespace, func(httpAddrs []string, grpcAddrs []string) {
			c.httpAddrs = httpAddrs
			c.grpcAddrs = grpcAddrs
			c.lbIndex = 0
			c.applyTargets()
		})
		c.discovery = d
		go d.Run(context.Background())
		go c.startLB()
	}
	return c
}

// NewClientWithDefaults creates a new Seata client with default configuration
func NewClientWithDefaults() *Client {
	return NewClient(DefaultConfig())
}

// StartTransaction creates a new global transaction
func (c *Client) StartTransaction(ctx context.Context, mode string, payload []byte) (*Transaction, error) {
	// Generate transaction ID
	gid := uuid.New().String()

	// Use gRPC if available, otherwise fall back to HTTP
	if c.grpcClient != nil && c.grpcClient.client != nil {
		return c.startTransactionGRPC(ctx, gid, mode, payload)
	}

	return c.startTransactionHTTP(ctx, gid, mode, payload)
}

// startTransactionHTTP creates a transaction via HTTP
func (c *Client) startTransactionHTTP(ctx context.Context, gid, mode string, payload []byte) (*Transaction, error) {
	// Prepare request - convert payload to integer array for JSON serialization
	payloadArray := bytesToIntArray(payload)

	req := map[string]interface{}{
		"gid":     gid,
		"mode":    mode,
		"payload": payloadArray,
	}

	// Make HTTP request
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/start")

	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to start transaction: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	// Parse response
	var result struct {
		GID string `json:"gid"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Create transaction object
	tx := &Transaction{
		client:   c,
		gid:      result.GID,
		mode:     mode,
		payload:  payload,
		branches: make([]*Branch, 0),
	}

	return tx, nil
}

// startTransactionGRPC creates a transaction via gRPC
func (c *Client) startTransactionGRPC(ctx context.Context, gid, mode string, payload []byte) (*Transaction, error) {
	resp, err := c.grpcClient.StartGlobal(ctx, gid, mode, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction via gRPC: %w", err)
	}

	// Create transaction object
	tx := &Transaction{
		client:   c,
		gid:      resp.Gid,
		mode:     mode,
		payload:  payload,
		branches: make([]*Branch, 0),
	}

	return tx, nil
}

// GetTransaction retrieves a transaction by its global ID
func (c *Client) GetTransaction(ctx context.Context, gid string) (*TransactionInfo, error) {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/api/tx/%s", gid))

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get transaction: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	var txInfo TransactionInfo
	if err := json.Unmarshal(resp.Body(), &txInfo); err != nil {
		return nil, fmt.Errorf("failed to parse transaction info: %w", err)
	}

	return &txInfo, nil
}

// ListTransactions retrieves a list of transactions with optional filtering
func (c *Client) ListTransactions(ctx context.Context, limit, offset int, status string) ([]*TransactionInfo, error) {
	req := c.httpClient.R().SetContext(ctx)

	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		req.SetQueryParam("offset", fmt.Sprintf("%d", offset))
	}
	if status != "" {
		req.SetQueryParam("status", status)
	}

	resp, err := req.Get("/api/tx")

	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list transactions: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	var transactions []*TransactionInfo
	if err := json.Unmarshal(resp.Body(), &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse transactions list: %w", err)
	}

	return transactions, nil
}

// Health checks the health of the Seata server
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get("/health")

	if err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	// Seata server returns plain text "ok" for health check
	body := resp.String()
	if body == "ok" {
		return &HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		}, nil
	}

	// Try to parse as JSON if not plain text
	var health HealthStatus
	if err := json.Unmarshal(resp.Body(), &health); err != nil {
		return nil, fmt.Errorf("failed to parse health status: %w", err)
	}

	return &health, nil
}

// Metrics retrieves Prometheus metrics from the server
func (c *Client) Metrics(ctx context.Context) (string, error) {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get("/metrics")

	if err != nil {
		return "", fmt.Errorf("failed to get metrics: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("failed to get metrics: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return resp.String(), nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	if c.discovery != nil {
		c.discovery.Stop()
	}
	if c.lbStop != nil {
		close(c.lbStop)
	}
	if c.grpcClient != nil {
		return c.grpcClient.Close()
	}
	return nil
}

// startLB starts a simple round-robin rotation across discovered endpoints
func (c *Client) startLB() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.lbStop:
			return
		case <-ticker.C:
			if len(c.httpAddrs) == 0 && len(c.grpcAddrs) == 0 {
				continue
			}
			c.lbIndex++
			c.applyTargets()
		}
	}
}

// applyTargets applies the current index to set HTTP BaseURL and gRPC target
func (c *Client) applyTargets() {
	if len(c.httpAddrs) > 0 {
		idx := c.lbIndex % len(c.httpAddrs)
		if idx < 0 {
			idx = 0
		}
		c.httpClient.SetBaseURL(c.httpAddrs[idx])
	}
	if len(c.grpcAddrs) > 0 {
		idx := c.lbIndex % len(c.grpcAddrs)
		if idx < 0 {
			idx = 0
		}
		_ = c.grpcClient.Close()
		c.grpcClient = NewGrpcClient(c.grpcAddrs[idx])
	}
}

// EtcdDiscovery watches endpoints in etcd and updates client targets
type EtcdDiscovery struct {
	endpoints []string
	namespace string
	onUpdate  func([]string, []string)
	stopCh    chan struct{}
}

func NewEtcdDiscovery(endpoints []string, namespace string, onUpdate func([]string, []string)) *EtcdDiscovery {
	if namespace == "" {
		namespace = "/seata"
	}
	return &EtcdDiscovery{endpoints: endpoints, namespace: namespace, onUpdate: onUpdate, stopCh: make(chan struct{})}
}

func (d *EtcdDiscovery) Run(ctx context.Context) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: d.endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		return
	}
	defer cli.Close()

	// initial fetch
	httpAddrs := d.fetch(cli, d.namespace+"/endpoints/http/")
	grpcAddrs := d.fetch(cli, d.namespace+"/endpoints/grpc/")
	if d.onUpdate != nil {
		d.onUpdate(httpAddrs, grpcAddrs)
	}

	// watch
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	wchHttp := cli.Watch(watchCtx, d.namespace+"/endpoints/http/", clientv3.WithPrefix())
	wchGrpc := cli.Watch(watchCtx, d.namespace+"/endpoints/grpc/", clientv3.WithPrefix())

	for {
		select {
		case <-d.stopCh:
			return
		case <-watchCtx.Done():
			return
		case <-wchHttp:
			httpAddrs = d.fetch(cli, d.namespace+"/endpoints/http/")
			if d.onUpdate != nil {
				d.onUpdate(httpAddrs, grpcAddrs)
			}
		case <-wchGrpc:
			grpcAddrs = d.fetch(cli, d.namespace+"/endpoints/grpc/")
			if d.onUpdate != nil {
				d.onUpdate(httpAddrs, grpcAddrs)
			}
		}
	}
}

func (d *EtcdDiscovery) Stop() { close(d.stopCh) }

func (d *EtcdDiscovery) fetch(cli *clientv3.Client, prefix string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	addrs := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}
	return addrs
}
