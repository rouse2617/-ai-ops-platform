package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// StdioConfig stdio 客户端配置
type StdioConfig struct {
	Name    string   `json:"name"`
	Command string   `json:"command"` // 可执行文件路径
	Args    []string `json:"args"`    // 命令行参数
	Env     []string `json:"env"`     // 环境变量
	Enabled bool     `json:"enabled"`
}

// StdioClient stdio 协议 MCP 客户端
type StdioClient struct {
	name    string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	mu      sync.Mutex
	reqID   int64
	pending map[int64]chan *JSONRPCResponse
}

// NewStdioClient 创建 stdio 客户端
func NewStdioClient(cfg StdioConfig) (*StdioClient, error) {
	cmd := exec.Command(cfg.Command, cfg.Args...)
	if len(cfg.Env) > 0 {
		cmd.Env = append(cmd.Env, cfg.Env...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdin 管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdout 管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动进程失败: %w", err)
	}

	client := &StdioClient{
		name:    cfg.Name,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		pending: make(map[int64]chan *JSONRPCResponse),
	}

	// 启动响应读取协程
	go client.readResponses()

	// 执行 MCP 初始化握手
	ctx := context.Background()
	if err := client.initialize(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("MCP 初始化失败: %w", err)
	}

	return client, nil
}

// initialize 执行 MCP 协议初始化
func (c *StdioClient) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    ClientCapability{},
		ClientInfo: ClientInfo{
			Name:    "ai-ops",
			Version: "1.0.0",
		},
	}

	resp, err := c.call(ctx, MethodInitialize, params)
	if err != nil {
		return err
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("解析初始化响应失败: %w", err)
	}

	// 发送 initialized 通知
	return c.notify(MethodInitialized, nil)
}

// notify 发送通知（无需响应）
func (c *StdioClient) notify(method string, params interface{}) error {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.mu.Lock()
	_, err = c.stdin.Write(append(data, '\n'))
	c.mu.Unlock()
	return err
}

// readResponses 读取响应
func (c *StdioClient) readResponses() {
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		c.mu.Lock()
		if ch, ok := c.pending[resp.ID]; ok {
			ch <- &resp
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
	}
}

// call 发送 JSON-RPC 请求
func (c *StdioClient) call(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	id := atomic.AddInt64(&c.reqID, 1)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 创建响应通道
	respCh := make(chan *JSONRPCResponse, 1)
	c.mu.Lock()
	c.pending[id] = respCh
	c.mu.Unlock()

	// 发送请求
	c.mu.Lock()
	_, err = c.stdin.Write(append(data, '\n'))
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 等待响应
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC 错误: %s", resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

// ListTools 列出工具
func (c *StdioClient) ListTools(ctx context.Context) ([]Tool, error) {
	resp, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []Tool `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// CallTool 调用工具
func (c *StdioClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	resp, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	var result ToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetHealth 健康检查
func (c *StdioClient) GetHealth(ctx context.Context) error {
	if c.cmd.ProcessState != nil && c.cmd.ProcessState.Exited() {
		return fmt.Errorf("进程已退出")
	}
	return nil
}

// Close 关闭客户端
func (c *StdioClient) Close() error {
	c.stdin.Close()
	return c.cmd.Process.Kill()
}

// Name 获取客户端名称
func (c *StdioClient) Name() string {
	return c.name
}
