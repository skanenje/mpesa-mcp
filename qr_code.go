package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QRCodeRequest represents the request for QR code generation
type QRCodeRequest struct {
	MerchantName string `json:"MerchantName"`
	RefNo        string `json:"RefNo"`
	Amount       int    `json:"Amount"`
	TrxCode      string `json:"TrxCode"`
	CPI          string `json:"CPI"`
	Size         string `json:"Size"`
}

// QRCodeResponse represents the response from QR code generation
type QRCodeResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	RequestID           string `json:"RequestID"`
	ResponseDescription string `json:"ResponseDescription"`
	QRCode              string `json:"QRCode"`
	ErrorCode           string `json:"errorCode,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}

// GenerateQRCode generates a QR code for M-Pesa payment
func (ctx *AppContext) GenerateQRCode(
	bgCtx context.Context,
	merchantName string,
	refNo string,
	amount int,
	trxCode string,
	cpi string,
) (*QRCodeResponse, error) {
	// Validate transaction code
	validTrxCodes := map[string]bool{
		"BG": true, // Buy Goods
		"WA": true, // Withdraw Cash
		"PB": true, // Paybill
		"SM": true, // Send Money (Mobile number)
		"SB": true, // Send to Business
	}
	
	if !validTrxCodes[trxCode] {
		return nil, fmt.Errorf("invalid transaction code: %s (must be BG, WA, PB, SM, or SB)", trxCode)
	}

	// Create request payload
	payload := QRCodeRequest{
		MerchantName: merchantName,
		RefNo:        refNo,
		Amount:       amount,
		TrxCode:      trxCode,
		CPI:          cpi,
		Size:         "300",
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/mpesa/qrcode/v1/generate", ctx.baseURL)
	req, err := http.NewRequestWithContext(bgCtx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ctx.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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
	var qrResp QRCodeResponse
	if err := json.Unmarshal(body, &qrResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if qrResp.ErrorCode != "" {
		return &qrResp, fmt.Errorf("QR code generation failed: %s - %s", qrResp.ErrorCode, qrResp.ErrorMessage)
	}

	if resp.StatusCode != http.StatusOK {
		return &qrResp, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return &qrResp, nil
}