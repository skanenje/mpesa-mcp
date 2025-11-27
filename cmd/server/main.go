package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mpesa-mcp/internal/config"
	"mpesa-mcp/internal/mcp"
	"mpesa-mcp/internal/mpesa"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
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

	// Initialize MCP server
	mcpServer := mcp.NewServer(mpesaClient)

	// Create SSE handler using SDK's built-in handler
	// This properly handles multiple concurrent sessions
	sseHandler := mcpsdk.NewSSEHandler(func(r *http.Request) *mcpsdk.Server {
		return mcpServer.GetMCPServer()
	})

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/sse", sseHandler)
	mux.Handle("/message", sseHandler)
	mux.HandleFunc("/callback", mpesaClient.ProcessCallback)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // No timeout for SSE
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 M-Pesa MCP Server starting on %s", addr)
		log.Printf("📡 SSE endpoint: http://localhost%s/sse", addr)
		log.Printf("💬 Message endpoint: http://localhost%s/message", addr)
		log.Printf("❤️  Health check: http://localhost%s/health", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"mpesa-mcp","transport":"sse"}`))
}
