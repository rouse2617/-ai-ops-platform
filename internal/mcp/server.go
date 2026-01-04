package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"ai-ops/internal/tool"
)

// Server MCP 服务端，暴露内置工具给外部 MCP 客户端
type Server struct {
	registry *tool.Registry
	info     ServerInfo
	mu       sync.RWMutex
}

// NewServer 创建 MCP 服务端
func NewServer(registry *tool.Registry) *Server {
	return &Server{
		registry: registry,
		info: ServerInfo{
			Name:    "ai-ops",
			Version: "1.0.0",
		},
	}
}

// HandleJSONRPC 处理 JSON-RPC 2.0 请求
func (s *Server) HandleJSONRPC(ctx context.Context, data []byte) ([]byte, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return s.errorResponse(0, -32700, "Parse error", nil)
	}

	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case MethodToolsList:
		return s.handleToolsList(req)
	case MethodToolsCall:
		return s.handleToolsCall(ctx, req)
	default:
		// 通知消息（无 ID）不需要响应
		if req.ID == 0 && req.Method == MethodInitialized {
			return nil, nil
		}
		return s.errorResponse(req.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize 处理初始化请求
func (s *Server) handleInitialize(req JSONRPCRequest) ([]byte, error) {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapability{
			Tools: &ToolsCapability{ListChanged: false},
		},
		ServerInfo: s.info,
	}
	return s.successResponse(req.ID, result)
}

// handleToolsList 处理工具列表请求
func (s *Server) handleToolsList(req JSONRPCRequest) ([]byte, error) {
	tools := s.registry.List()
	mcpTools := make([]ToolInfo, 0, len(tools))

	for _, t := range tools {
		if !s.registry.IsEnabled(t.Name()) {
			continue
		}
		mcpTools = append(mcpTools, ToolInfo{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: s.buildInputSchema(t.Parameters()),
		})
	}

	result := ToolsListResult{Tools: mcpTools}
	return s.successResponse(req.ID, result)
}

// handleToolsCall 处理工具调用请求
func (s *Server) handleToolsCall(ctx context.Context, req JSONRPCRequest) ([]byte, error) {
	// 解析参数
	paramsData, err := json.Marshal(req.Params)
	if err != nil {
		return s.errorResponse(req.ID, -32602, "Invalid params", nil)
	}

	var params ToolCallParams
	if err := json.Unmarshal(paramsData, &params); err != nil {
		return s.errorResponse(req.ID, -32602, "Invalid params", nil)
	}

	// 执行工具
	toolCtx := &tool.Context{}
	result, err := s.registry.Execute(toolCtx, params.Name, params.Arguments)
	if err != nil {
		return s.successResponse(req.ID, ToolCallResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("错误: %s", err.Error())}},
			IsError: true,
		})
	}

	// 转换结果
	text := ""
	if result.Message != "" {
		text = result.Message
	}
	if result.Data != nil {
		dataJSON, _ := json.MarshalIndent(result.Data, "", "  ")
		if text != "" {
			text += "\n"
		}
		text += string(dataJSON)
	}

	isError := !result.Success || result.Error != ""

	return s.successResponse(req.ID, ToolCallResult{
		Content: []Content{{Type: "text", Text: text}},
		IsError: isError,
	})
}

// buildInputSchema 构建工具输入 Schema
func (s *Server) buildInputSchema(params []tool.Parameter) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, p := range params {
		prop := map[string]interface{}{
			"type":        p.Type,
			"description": p.Description,
		}
		if len(p.Enum) > 0 {
			prop["enum"] = p.Enum
		}
		if p.Default != nil {
			prop["default"] = p.Default
		}
		properties[p.Name] = prop

		if p.Required {
			required = append(required, p.Name)
		}
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// successResponse 构建成功响应
func (s *Server) successResponse(id int64, result interface{}) ([]byte, error) {
	resultData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultData,
	}
	return json.Marshal(resp)
}

// errorResponse 构建错误响应
func (s *Server) errorResponse(id int64, code int, message string, data interface{}) ([]byte, error) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return json.Marshal(resp)
}
