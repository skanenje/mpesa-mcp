package mpesa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RefreshToken gets a new access token from the Daraja API
func (c *Client) RefreshToken(ctx context.Context) error {
	// Create basic auth header
	auth := c.config.ConsumerKey + ":" + c.config.ConsumerSec
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Create request
	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update token (thread-safe)
	expiry := time.Now().Add(55 * time.Minute)
	c.setAccessToken(tokenResp.AccessToken, expiry)

	return nil
}
