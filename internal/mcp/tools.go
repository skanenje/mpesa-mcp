package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// STKPushInput defines the input for STK Push
type STKPushInput struct {
	Amount      int    `json:"amount" jsonschema:"required,description=Amount to be paid in KES"`
	PhoneNumber string `json:"phone_number" jsonschema:"required,description=Phone number of customer (format: 254XXXXXXXXX or 0XXXXXXXXX)"`
}

// QRCodeInput defines the input for QR code generation
type QRCodeInput struct {
	MerchantName string `json:"merchant_name" jsonschema:"required,description=Name of the merchant/business"`
	RefNo        string `json:"ref_no" jsonschema:"required,description=Transaction reference number"`
	Amount       int    `json:"amount" jsonschema:"required,description=Amount to be paid in KES"`
	TrxCode      string `json:"trx_code" jsonschema:"required,description=Transaction code (BG=Buy Goods WA=Withdraw PB=Paybill SM=Send Money SB=Send to Business)"`
	CPIdentifier string `json:"cp_identifier" jsonschema:"required,description=Credit party identifier (till number paybill or phone number)"`
}

// registerTools registers all M-Pesa tools with the MCP server
func (s *Server) registerTools() {
	// STK Push tool
	mcpsdk.AddTool(
		s.mcp,
		&mcpsdk.Tool{
			Name:        "stk_push",
			Description: "Initiate an M-Pesa STK Push payment request. This prompts the customer to authorize payment on their mobile device.",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, input STKPushInput) (*mcpsdk.CallToolResult, map[string]interface{}, error) {
			response, err := s.mpesa.InitiateSTKPush(ctx, input.Amount, input.PhoneNumber)
			if err != nil {
				return nil, nil, fmt.Errorf("STK Push failed: %w", err)
			}

			jsonData, _ := json.Marshal(response)
			var resultMap map[string]interface{}
			json.Unmarshal(jsonData, &resultMap)

			return &mcpsdk.CallToolResult{
				Content: []interface{}{
					&mcpsdk.TextContent{
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
	mcpsdk.AddTool(
		s.mcp,
		&mcpsdk.Tool{
			Name:        "generate_qr_code",
			Description: "Generate an M-Pesa QR code that customers can scan to make payment.",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, input QRCodeInput) (*mcpsdk.CallToolResult, map[string]interface{}, error) {
			response, err := s.mpesa.GenerateQRCode(
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

			jsonData, _ := json.Marshal(response)
			var resultMap map[string]interface{}
			json.Unmarshal(jsonData, &resultMap)

			return &mcpsdk.CallToolResult{
				Content: []interface{}{
					&mcpsdk.TextContent{
						Type: "text",
						Text: fmt.Sprintf("QR Code generated successfully!\n\nRequest ID: %s\nQR Code (Base64): %s...",
							response.RequestID,
							response.QRCode[:min(50, len(response.QRCode))],
						),
					},
				},
			}, resultMap, nil
		},
	)

	// Token status tool
	mcpsdk.AddTool(
		s.mcp,
		&mcpsdk.Tool{
			Name:        "get_token_status",
			Description: "Get the current access token status and expiry time.",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, input struct{}) (*mcpsdk.CallToolResult, map[string]interface{}, error) {
			status := map[string]interface{}{
				"has_token":  s.mpesa.GetAccessToken() != "",
				"expires_at": s.mpesa.GetTokenExpiry().Format("2006-01-02 15:04:05"),
				"is_valid":   s.mpesa.IsTokenValid(),
			}

			return &mcpsdk.CallToolResult{
				Content: []interface{}{
					&mcpsdk.TextContent{
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
