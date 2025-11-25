package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TokenResponse represents the OAuth token response from Daraja API
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// refreshToken gets a new access token from the Daraja API
func (ctx *AppContext) refreshToken(bgCtx context.Context) error {
	// Create basic auth header
	auth := ctx.consumerKey + ":" + ctx.consumerSec
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Create request
	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", ctx.baseURL)
	req, err := http.NewRequestWithContext(bgCtx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

	// Update context
	ctx.accessToken = tokenResp.AccessToken
	ctx.tokenExpiry = time.Now().Add(55 * time.Minute) // Set expiry slightly before actual

	return nil
}