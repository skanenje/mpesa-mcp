package mpesa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mpesa-mcp/internal/utils"
)

// InitiateSTKPush initiates an STK Push payment request
func (c *Client) InitiateSTKPush(ctx context.Context, amount int, phoneNumber string) (*STKPushResponse, error) {
	return c.InitiateSTKPushWithOptions(ctx, amount, phoneNumber, "", "")
}

// InitiateSTKPushWithOptions initiates an STK Push payment request with optional parameters
func (c *Client) InitiateSTKPushWithOptions(ctx context.Context, amount int, phoneNumber, accountRef, transDesc string) (*STKPushResponse, error) {
	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Generate password: Base64(ShortCode + Passkey + Timestamp)
	passwordStr := c.config.BusinessCode + c.config.Passkey + timestamp
	password := base64.StdEncoding.EncodeToString([]byte(passwordStr))

	// Format phone number
	formattedPhone := utils.FormatPhoneNumber(phoneNumber)

	// Use provided values or defaults
	if accountRef == "" {
		accountRef = c.config.AccountRef
	}
	if transDesc == "" {
		transDesc = "Payment for goods/services"
	}

	// Validate field lengths per M-Pesa documentation
	if len(accountRef) > 12 {
		accountRef = accountRef[:12]
	}
	if len(transDesc) > 13 {
		transDesc = transDesc[:13]
	}

	// Create request payload
	payload := STKPushRequest{
		BusinessShortCode: c.config.BusinessCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            formattedPhone,
		PartyB:            c.config.BusinessCode,
		PhoneNumber:       formattedPhone,
		CallBackURL:       c.config.CallbackURL,
		AccountReference:  accountRef,
		TransactionDesc:   transDesc,
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.GetAccessToken())
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if stkResp.ErrorCode != "" {
		return &stkResp, fmt.Errorf("STK Push failed: %s - %s", stkResp.ErrorCode, stkResp.ErrorMessage)
	}

	if resp.StatusCode != http.StatusOK {
		return &stkResp, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, stkResp.ResponseDescription)
	}

	return &stkResp, nil
}
