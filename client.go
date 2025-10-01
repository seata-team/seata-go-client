package seata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// Client represents a Seata client for distributed transaction management
type Client struct {
	httpClient *resty.Client
	grpcClient *GrpcClient
	config     *Config
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
	}
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

	return &Client{
		httpClient: httpClient,
		grpcClient: grpcClient,
		config:     config,
	}
}

// NewClientWithDefaults creates a new Seata client with default configuration
func NewClientWithDefaults() *Client {
	return NewClient(DefaultConfig())
}

// StartTransaction creates a new global transaction
func (c *Client) StartTransaction(ctx context.Context, mode string, payload []byte) (*Transaction, error) {
	// Generate transaction ID
	gid := uuid.New().String()

	// Encode payload to base64
	encodedPayload := base64.StdEncoding.EncodeToString(payload)

	// Prepare request
	req := map[string]interface{}{
		"gid":     gid,
		"mode":    mode,
		"payload": encodedPayload,
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
	if c.grpcClient != nil {
		return c.grpcClient.Close()
	}
	return nil
}
