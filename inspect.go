package main

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	var _ mcp.JSONRPCMessage = &mcp.JSONRPCRequest{}
	fmt.Println("JSONRPCRequest implements JSONRPCMessage")

	var _ mcp.JSONRPCMessage = &mcp.JSONRPCResponse{}
	fmt.Println("JSONRPCResponse implements JSONRPCMessage")

	// var _ mcp.JSONRPCMessage = &mcp.JSONRPCError{}
	// fmt.Println("JSONRPCError implements JSONRPCMessage")
}
