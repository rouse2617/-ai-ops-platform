package mcp

import "context"

// MCPClient MCP 客户端接口
type MCPClient interface {
	Name() string
	ListTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
	GetHealth(ctx context.Context) error
	Close() error
}
