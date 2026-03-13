package mcputil

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ErrorResult returns an MCP tool result indicating an error.
func ErrorResult(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}, nil, nil
}

// TextResult returns an MCP tool result with plain text content.
func TextResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}
