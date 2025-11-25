package mcp

import (
	"mpesa-mcp/internal/mpesa"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHandler creates a new MCP server (alias for NewServer for compatibility)
func NewHandler(mpesaClient *mpesa.Client) *mcpsdk.Server {
	server := NewServer(mpesaClient)
	return server.GetMCPServer()
}
