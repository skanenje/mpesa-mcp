package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPrompts registers all M-Pesa prompts with the MCP server
func (s *Server) registerPrompts() {
	s.registerSTKPushPrompt()
	s.registerQRCodePrompt()
	s.registerPaymentAssistancePrompt()
}

func (s *Server) registerSTKPushPrompt() {
	s.mcp.AddPrompts(
		&mcpsdk.ServerPrompt{
			Prompt: &mcpsdk.Prompt{
				Name:        "stk_push_prompt",
				Description: "Generate a prompt for initiating an M-Pesa STK Push payment request",
				Arguments: []*mcpsdk.PromptArgument{
					{
						Name:        "phone_number",
						Description: "The phone number of the customer",
						Required:    true,
					},
					{
						Name:        "amount",
						Description: "The amount to be paid in KES",
						Required:    true,
					},
					{
						Name:        "purpose",
						Description: "The purpose of the payment",
						Required:    true,
					},
				},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.GetPromptParams) (*mcpsdk.GetPromptResult, error) {
				args := params.Arguments
				phoneNumber := args["phone_number"]
				amount := args["amount"]
				purpose := args["purpose"]

				promptText := fmt.Sprintf(
					"I want you to initiate an M-Pesa STK Push payment request with the following details:\n\n"+
						"- Customer Phone Number: %s\n"+
						"- Amount: KES %s\n"+
						"- Purpose: %s\n\n"+
						"Please use the stk_push tool to process this payment request.",
					phoneNumber, amount, purpose,
				)

				return &mcpsdk.GetPromptResult{
					Description: "STK Push payment request prompt",
					Messages: []*mcpsdk.PromptMessage{
						{
							Role: "user",
							Content: &mcpsdk.TextContent{
								Text: promptText,
							},
						},
					},
				}, nil
			},
		},
	)
}

func (s *Server) registerQRCodePrompt() {
	s.mcp.AddPrompts(
		&mcpsdk.ServerPrompt{
			Prompt: &mcpsdk.Prompt{
				Name:        "generate_qr_code_prompt",
				Description: "Generate a prompt for creating an M-Pesa QR code payment request",
				Arguments: []*mcpsdk.PromptArgument{
					{
						Name:        "merchant_name",
						Description: "Name of the merchant/business",
						Required:    true,
					},
					{
						Name:        "amount",
						Description: "Amount to be paid in KES",
						Required:    true,
					},
					{
						Name:        "transaction_type",
						Description: "Type of transaction (BG=Buy Goods, WA=Wallet, PB=Paybill, SM=Send Money, SB=Send to Business)",
						Required:    true,
					},
					{
						Name:        "identifier",
						Description: "The recipient identifier (till number, paybill, phone number)",
						Required:    true,
					},
					{
						Name:        "reference",
						Description: "Transaction reference number (optional)",
						Required:    false,
					},
				},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.GetPromptParams) (*mcpsdk.GetPromptResult, error) {
				args := params.Arguments
				merchantName := args["merchant_name"]
				amount := args["amount"]
				transactionType := args["transaction_type"]
				identifier := args["identifier"]
				reference := args["reference"]

				if reference == "" {
					reference = "QR_PAYMENT"
				}

				transactionTypes := map[string]string{
					"BG": "Buy Goods",
					"WA": "Wallet",
					"PB": "Paybill",
					"SM": "Send Money",
					"SB": "Send to Business",
				}

				trxDescription := transactionTypes[transactionType]
				if trxDescription == "" {
					trxDescription = transactionType
				}

				promptText := fmt.Sprintf(
					"I want to generate an M-Pesa QR code with the following details:\n\n"+
						"- Merchant/Business Name: %s\n"+
						"- Amount: KES %s\n"+
						"- Transaction Type: %s (%s)\n"+
						"- Recipient Identifier: %s\n"+
						"- Reference Number: %s\n\n"+
						"Please use the generate_qr_code tool to create this QR code that customers can scan to make payment.",
					merchantName, amount, trxDescription, transactionType, identifier, reference,
				)

				return &mcpsdk.GetPromptResult{
					Description: "QR code generation prompt",
					Messages: []*mcpsdk.PromptMessage{
						{
							Role: "user",
							Content: &mcpsdk.TextContent{
								Text: promptText,
							},
						},
					},
				}, nil
			},
		},
	)
}

func (s *Server) registerPaymentAssistancePrompt() {
	s.mcp.AddPrompts(
		&mcpsdk.ServerPrompt{
			Prompt: &mcpsdk.Prompt{
				Name:        "payment_assistance",
				Description: "Get help with M-Pesa payment integration",
				Arguments:   []*mcpsdk.PromptArgument{},
			},
			Handler: func(ctx context.Context, session *mcpsdk.ServerSession, params *mcpsdk.GetPromptParams) (*mcpsdk.GetPromptResult, error) {
				promptText := `I need assistance with M-Pesa payment integration. Here's what this MCP server can do:

**Available Tools:**

1. **stk_push** - Initiate a payment request to a customer's phone
   - Sends a push notification to the customer's M-Pesa app
   - Customer enters their M-Pesa PIN to authorize payment
   - Requires: amount (in KES) and phone number (254XXXXXXXXX format)

2. **generate_qr_code** - Create a QR code for payment
   - Customers scan the QR code with their M-Pesa app
   - Requires: merchant name, amount, transaction type, and identifier
   - Transaction types: BG (Buy Goods), PB (Paybill), SM (Send Money), etc.

3. **get_token_status** - Check the OAuth token status
   - Useful for debugging authentication issues

**Example Scenarios:**
- "Charge customer 254712345678 KES 1000 for order #12345"
- "Generate a Buy Goods QR code for KES 500 at My Store, till number 123456"

How can I help you with M-Pesa payments today?`

				return &mcpsdk.GetPromptResult{
					Description: "M-Pesa payment assistance",
					Messages: []*mcpsdk.PromptMessage{
						{
							Role: "user",
							Content: &mcpsdk.TextContent{
								Text: promptText,
							},
						},
					},
				}, nil
			},
		},
	)
}
