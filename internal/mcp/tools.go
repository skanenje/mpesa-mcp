package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
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
	s.mcp.AddTools(
		&mcpsdk.ServerTool{
			Tool: &mcpsdk.Tool{
				Name:        "stk_push",
				Description: "Initiate an M-Pesa STK Push payment request. This prompts the customer to authorize payment on their mobile device.",
				InputSchema: &jsonschema.Schema{
					Type: "object",
					Properties: map[string]*jsonschema.Schema{
						"amount": {
							Type:        "integer",
							Description: "Amount to be paid in KES",
						},
						"phone_number": {
							Type:        "string",
							Description: "Phone number of customer (format: 254XXXXXXXXX or 0XXXXXXXXX)",
						},
					},
					Required: []string{"amount", "phone_number"},
				},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.CallToolParamsFor[map[string]interface{}]) (*mcpsdk.CallToolResultFor[interface{}], error) {
				log.Printf("[Tool:stk_push] Called with params: %+v", params.Arguments)

				args := params.Arguments
				amountFloat, _ := args["amount"].(float64)
				amount := int(amountFloat)
				phoneNumber, _ := args["phone_number"].(string)

				log.Printf("[Tool:stk_push] Initiating STK Push - Amount: %d, Phone: %s", amount, phoneNumber)

				response, err := s.mpesa.InitiateSTKPush(ctx, amount, phoneNumber)
				if err != nil {
					log.Printf("[Tool:stk_push] Failed: %v", err)
					return nil, fmt.Errorf("STK Push failed: %w", err)
				}

				log.Printf("[Tool:stk_push] Success - MerchantRequestID: %s, CheckoutRequestID: %s", response.MerchantRequestID, response.CheckoutRequestID)

				jsonData, _ := json.Marshal(response)
				var resultMap map[string]interface{}
				json.Unmarshal(jsonData, &resultMap)

				return &mcpsdk.CallToolResultFor[interface{}]{
					Content: []mcpsdk.Content{
						&mcpsdk.TextContent{
							Text: fmt.Sprintf("STK Push initiated successfully!\n\nMerchant Request ID: %s\nCheckout Request ID: %s\nCustomer Message: %s",
								response.MerchantRequestID,
								response.CheckoutRequestID,
								response.CustomerMessage,
							),
						},
					},
				}, nil
			},
		},
	)

	// QR Code generation tool
	s.mcp.AddTools(
		&mcpsdk.ServerTool{
			Tool: &mcpsdk.Tool{
				Name:        "generate_qr_code",
				Description: "Generate an M-Pesa QR code that customers can scan to make payment.",
				InputSchema: &jsonschema.Schema{
					Type: "object",
					Properties: map[string]*jsonschema.Schema{
						"merchant_name": {
							Type:        "string",
							Description: "Name of the merchant/business",
						},
						"ref_no": {
							Type:        "string",
							Description: "Transaction reference number",
						},
						"amount": {
							Type:        "integer",
							Description: "Amount to be paid in KES",
						},
						"trx_code": {
							Type:        "string",
							Description: "Transaction code (BG=Buy Goods WA=Withdraw PB=Paybill SM=Send Money SB=Send to Business)",
						},
						"cp_identifier": {
							Type:        "string",
							Description: "Credit party identifier (till number paybill or phone number)",
						},
					},
					Required: []string{"merchant_name", "ref_no", "amount", "trx_code", "cp_identifier"},
				},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.CallToolParamsFor[map[string]interface{}]) (*mcpsdk.CallToolResultFor[interface{}], error) {
				log.Printf("[Tool:generate_qr_code] Called with params: %+v", params.Arguments)

				args := params.Arguments
				merchantName, _ := args["merchant_name"].(string)
				refNo, _ := args["ref_no"].(string)
				amountFloat, _ := args["amount"].(float64)
				amount := int(amountFloat)
				trxCode, _ := args["trx_code"].(string)
				cpIdentifier, _ := args["cp_identifier"].(string)

				log.Printf("[Tool:generate_qr_code] Generating QR - Merchant: %s, Amount: %d", merchantName, amount)

				response, err := s.mpesa.GenerateQRCode(
					ctx,
					merchantName,
					refNo,
					amount,
					trxCode,
					cpIdentifier,
				)
				if err != nil {
					log.Printf("[Tool:generate_qr_code] Failed: %v", err)
					return nil, fmt.Errorf("QR code generation failed: %w", err)
				}

				log.Printf("[Tool:generate_qr_code] Success - RequestID: %s", response.RequestID)

				jsonData, _ := json.Marshal(response)
				var resultMap map[string]interface{}
				json.Unmarshal(jsonData, &resultMap)

				return &mcpsdk.CallToolResultFor[interface{}]{
					Content: []mcpsdk.Content{
						&mcpsdk.TextContent{
							Text: fmt.Sprintf("QR Code generated successfully!\n\nRequest ID: %s\nQR Code (Base64): %s...",
								response.RequestID,
								response.QRCode[:min(50, len(response.QRCode))],
							),
						},
					},
				}, nil
			},
		},
	)

	// Token status tool
	s.mcp.AddTools(
		&mcpsdk.ServerTool{
			Tool: &mcpsdk.Tool{
				Name:        "get_token_status",
				Description: "Get the current access token status and expiry time.",
				InputSchema: &jsonschema.Schema{
					Type: "object",
				},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.CallToolParamsFor[map[string]interface{}]) (*mcpsdk.CallToolResultFor[interface{}], error) {
				log.Printf("[Tool:get_token_status] Called")

				status := map[string]interface{}{
					"has_token":  s.mpesa.GetAccessToken() != "",
					"expires_at": s.mpesa.GetTokenExpiry().Format("2006-01-02 15:04:05"),
					"is_valid":   s.mpesa.IsTokenValid(),
				}

				log.Printf("[Tool:get_token_status] Status: %+v", status)

				return &mcpsdk.CallToolResultFor[interface{}]{
					Content: []mcpsdk.Content{
						&mcpsdk.TextContent{
							Text: fmt.Sprintf("Token Status:\n- Has Token: %v\n- Expires At: %s\n- Is Valid: %v",
								status["has_token"],
								status["expires_at"],
								status["is_valid"],
							),
						},
					},
				}, nil
			},
		},
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
