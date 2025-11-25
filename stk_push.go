package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// STKPushRequest represents the request payload for STK Push
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse represents the response from STK Push API
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
	ErrorCode           string `json:"errorCode,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}

// InitiateSTKPush initiates an STK Push payment request
func (ctx *AppContext) InitiateSTKPush(bgCtx context.Context, amount int, phoneNumber string) (*STKPushResponse, error) {
	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Generate password: Base64(ShortCode + Passkey + Timestamp)
	passwordStr := ctx.businessCode + ctx.passkey + timestamp
	password := base64.StdEncoding.EncodeToString([]byte(passwordStr))

	// Format phone number (remove leading 0 if present, ensure it starts with 254)
	formattedPhone := formatPhoneNumber(phoneNumber)

	// Create request payload
	payload := STKPushRequest{
		BusinessShortCode: ctx.businessCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            formattedPhone,
		PartyB:            ctx.businessCode,
		PhoneNumber:       formattedPhone,
		CallBackURL:       ctx.callbackURL,
		AccountReference:  ctx.accountRef,
		TransactionDesc:   "Payment for goods/services",
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", ctx.baseURL)
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

// formatPhoneNumber ensures phone number is in format 254XXXXXXXXX
func formatPhoneNumber(phone string) string {
	// Remove any spaces or special characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "+", "")
	
	// If starts with 0, replace with 254
	if strings.HasPrefix(phone, "0") {
		return "254" + phone[1:]
	}
	
	// If doesn't start with 254, add it
	if !strings.HasPrefix(phone, "254") {
		return "254" + phone
	}
	
	return phone
}