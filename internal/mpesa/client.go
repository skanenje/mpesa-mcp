package mpesa

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"mpesa-mcp/internal/config"
)

// Client handles all M-Pesa API interactions
type Client struct {
	config      *config.Config
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// NewClient creates a new M-Pesa API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken returns the current access token (thread-safe)
func (c *Client) GetAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

// setAccessToken updates the access token (thread-safe)
func (c *Client) setAccessToken(token string, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
	c.tokenExpiry = expiry
}

// IsTokenValid checks if the current token is still valid
func (c *Client) IsTokenValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken != "" && time.Now().Before(c.tokenExpiry)
}

// GetTokenExpiry returns when the token expires
func (c *Client) GetTokenExpiry() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tokenExpiry
}

// StartTokenRefresh begins the token refresh loop
func (c *Client) StartTokenRefresh(ctx context.Context) {
	// Get initial token
	if err := c.RefreshToken(ctx); err != nil {
		log.Printf("Failed to get initial access token: %v", err)
		return
	}

	// Refresh every 50 minutes (token expires in 60)
	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.RefreshToken(ctx); err != nil {
				log.Printf("Failed to refresh token: %v", err)
			}
		}
	}
}
