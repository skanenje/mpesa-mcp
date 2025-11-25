package mcp

import (
	"context"

	"mpesa-mcp/internal/mpesa"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with M-Pesa functionality
type Server struct {
	mcp    *mcpsdk.Server
	mpesa  *mpesa.Client
}

// NewServer creates a new MCP server with M-Pesa integration
func NewServer(mpesaClient *mpesa.Client) *Server {
	mcpServer := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "mpesa-mcp",
			Version: "v1.0.0",
		},
		nil,
	)

	s := &Server{
		mcp:   mcpServer,
		mpesa: mpesaClient,
	}

	// Register tools and prompts
	s.registerTools()
	s.registerPrompts()

	return s
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcpsdk.StdioTransport{})
}
