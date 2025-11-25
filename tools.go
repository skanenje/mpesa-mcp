package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// STKPushInput defines the input for STK Push
type STKPushInput struct {
	Amount      int    `json:"amount" jsonschema:"required,description=Amount to be paid in KES"`
	PhoneNumber string `json:"phone_number" jsonschema:"required,description=Phone number of customer (format: 254XXXXXXXXX or 0XXXXXXXXX)"`
}

// QRCodeInput defines the input for QR code generation
type QRCodeInput struct {
	MerchantName  string `json:"merchant_name" jsonschema:"required,description=Name of the merchant/business"`
	RefNo         string `json:"ref_no" jsonschema:"required,description=Transaction reference number"`
	Amount        int    `json:"amount" jsonschema:"required,description=Amount to be paid in KES"`
	TrxCode       string `json:"trx_code" jsonschema:"required,description=Transaction code (BG=Buy Goods WA=Withdraw PB=Paybill SM=Send Money SB=Send to Business)"`
	CPIdentifier  string `json:"cp_identifier" jsonschema:"required,description=Credit party identifier (till number paybill or phone number)"`
}

// registerMpesaTools registers all M-Pesa tools with the MCP server
func registerMpesaTools(server *mcp.Server, appCtx *AppContext) {
	// STK Push tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "stk_push",
			Description: "Initiate an M-Pesa STK Push payment request. This prompts the customer to authorize payment on their mobile device.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input STKPushInput) (*mcp.CallToolResult, map[string]interface{}, error) {
			// Initiate STK Push
			response, err := appCtx.InitiateSTKPush(ctx, input.Amount, input.PhoneNumber)
			if err != nil {
				return nil, nil, fmt.Errorf("STK Push failed: %w", err)
			}

			// Convert response to map
			jsonData, _ := json.Marshal(response)
			var resultMap map[string]interface{}
			json.Unmarshal(jsonData, &resultMap)

			return &mcp.CallToolResult{
				Content: []interface{}{
					&mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("STK Push initiated successfully!\n\nMerchant Request ID: %s\nCheckout Request ID: %s\nCustomer Message: %s",
							response.MerchantRequestID,
							response.CheckoutRequestID,
							response.CustomerMessage,
						),
					},
				},
			}, resultMap, nil
		},
	)

	// QR Code generation tool
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "generate_qr_code",
			Description: "Generate an M-Pesa QR code that customers can scan to make payment.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input QRCodeInput) (*mcp.CallToolResult, map[string]interface{}, error) {
			// Generate QR code
			response, err := appCtx.GenerateQRCode(
				ctx,
				input.MerchantName,
				input.RefNo,
				input.Amount,
				input.TrxCode,
				input.CPIdentifier,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("QR code generation failed: %w", err)
			}

			// Convert response to map
			jsonData, _ := json.Marshal(response)
			var resultMap map[string]interface{}
			json.Unmarshal(jsonData, &resultMap)

			return &mcp.CallToolResult{
				Content: []interface{}{
					&mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("QR Code generated successfully!\n\nRequest ID: %s\nQR Code (Base64): %s...",
							response.RequestID,
							response.QRCode[:50], // Show first 50 chars
						),
					},
				},
			}, resultMap, nil
		},
	)

	// Token status tool (helpful for debugging)
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_token_status",
			Description: "Get the current access token status and expiry time.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, map[string]interface{}, error) {
			status := map[string]interface{}{
				"has_token": appCtx.accessToken != "",
				"expires_at": appCtx.tokenExpiry.Format("2006-01-02 15:04:05"),
				"is_valid": appCtx.tokenExpiry.After(time.Now()),
			}

			return &mcp.CallToolResult{
				Content: []interface{}{
					&mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Token Status:\n- Has Token: %v\n- Expires At: %s\n- Is Valid: %v",
							status["has_token"],
							status["expires_at"],
							status["is_valid"],
						),
					},
				},
			}, status, nil
		},
	)
}