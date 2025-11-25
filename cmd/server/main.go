package main

import (
	"context"
	"log"

	"mpesa-mcp/internal/config"
	"mpesa-mcp/internal/mcp"
	"mpesa-mcp/internal/mpesa"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize M-Pesa client
	mpesaClient := mpesa.NewClient(cfg)

	// Start token refresh in background
	ctx := context.Background()
	go mpesaClient.StartTokenRefresh(ctx)

	// Initialize and run MCP server
	server := mcp.NewServer(mpesaClient)
	if err := server.Run(ctx); err != nil {
		log.Fatal("Server error:", err)
	}
}
