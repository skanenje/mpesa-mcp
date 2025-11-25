package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AppContext holds the application state
type AppContext struct {
	accessToken  string
	tokenExpiry  time.Time
	consumerKey  string
	consumerSec  string
	baseURL      string
	businessCode string
	passkey      string
	callbackURL  string
	accountRef   string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize context
	ctx := &AppContext{
		consumerKey:  os.Getenv("MPESA_CONSUMER_KEY"),
		consumerSec:  os.Getenv("MPESA_CONSUMER_SECRET"),
		baseURL:      os.Getenv("BASE_URL"),
		businessCode: os.Getenv("BUSINESS_SHORTCODE"),
		passkey:      os.Getenv("PASSKEY"),
		callbackURL:  os.Getenv("CALLBACK_URL"),
		accountRef:   os.Getenv("ACCOUNT_REFERENCE"),
	}

	// Validate required environment variables
	if err := ctx.validate(); err != nil {
		log.Fatal(err)
	}

	// Get initial access token
	if err := ctx.refreshToken(context.Background()); err != nil {
		log.Fatal("Failed to get initial access token:", err)
	}

	// Start token refresh goroutine
	go ctx.tokenRefreshLoop(context.Background())

	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mpesa-mcp",
			Version: "v1.0.0",
		},
		nil,
	)

	// Register tools
	registerMpesaTools(server, ctx)

	// Register prompts
	registerMpesaPrompts(server)

	// Run server over stdio
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func (ctx *AppContext) validate() error {
	if ctx.consumerKey == "" || ctx.consumerSec == "" {
		return fmt.Errorf("MPESA_CONSUMER_KEY and MPESA_CONSUMER_SECRET are required")
	}
	if ctx.baseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	if ctx.businessCode == "" || ctx.passkey == "" {
		return fmt.Errorf("BUSINESS_SHORTCODE and PASSKEY are required for STK Push")
	}
	return nil
}

func (ctx *AppContext) tokenRefreshLoop(bgCtx context.Context) {
	ticker := time.NewTicker(50 * time.Minute) // Refresh every 50 minutes (token expires in 60)
	defer ticker.Stop()

	for {
		select {
		case <-bgCtx.Done():
			return
		case <-ticker.C:
			if err := ctx.refreshToken(bgCtx); err != nil {
				log.Printf("Failed to refresh token: %v", err)
			}
		}
	}
}