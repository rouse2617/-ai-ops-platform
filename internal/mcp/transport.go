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

// Transport MCP 传输层接口
type Transport interface {
	// Send 发送请求并等待响应
	Send(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error)
	// SendNotification 发送通知（无需响应）
	SendNotification(ctx context.Context, req *JSONRPCRequest) error
	// Close 关闭传输层
	Close() error
}

// StdioTransport 基于 stdio 的传输层
type StdioTransport struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	scanner  *bufio.Scanner
	mu       sync.Mutex
	closed   bool
	pending  map[int64]chan *JSONRPCResponse
	pendingMu sync.Mutex
	nextID   int64
}

// NewStdioTransport 创建 stdio 传输层
func NewStdioTransport(command string, args []string, env []string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = env
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdin 管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("创建 stdout 管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("启动进程失败: %w", err)
	}

	t := &StdioTransport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		scanner: bufio.NewScanner(stdout),
		pending: make(map[int64]chan *JSONRPCResponse),
	}

	// 启动响应读取协程
	go t.readLoop()

	return t, nil
}

// readLoop 读取响应循环
func (t *StdioTransport) readLoop() {
	for t.scanner.Scan() {
		line := t.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		t.pendingMu.Lock()
		if ch, ok := t.pending[resp.ID]; ok {
			ch <- &resp
			delete(t.pending, resp.ID)
		}
		t.pendingMu.Unlock()
	}
}

// Send 发送请求并等待响应
func (t *StdioTransport) Send(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, fmt.Errorf("传输层已关闭")
	}

	// 分配请求 ID
	id := atomic.AddInt64(&t.nextID, 1)
	req.ID = id
	req.JSONRPC = "2.0"

	// 创建响应通道
	respCh := make(chan *JSONRPCResponse, 1)
	t.pendingMu.Lock()
	t.pending[id] = respCh
	t.pendingMu.Unlock()

	// 发送请求
	data, err := json.Marshal(req)
	if err != nil {
		t.mu.Unlock()
		t.pendingMu.Lock()
		delete(t.pending, id)
		t.pendingMu.Unlock()
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		t.mu.Unlock()
		t.pendingMu.Lock()
		delete(t.pending, id)
		t.pendingMu.Unlock()
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	t.mu.Unlock()

	// 等待响应
	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		t.pendingMu.Lock()
		delete(t.pending, id)
		t.pendingMu.Unlock()
		return nil, ctx.Err()
	}
}

// SendNotification 发送通知
func (t *StdioTransport) SendNotification(ctx context.Context, req *JSONRPCRequest) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("传输层已关闭")
	}

	req.JSONRPC = "2.0"
	req.ID = 0 // 通知没有 ID

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化通知失败: %w", err)
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	return nil
}

// Close 关闭传输层
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	t.stdin.Close()
	t.stdout.Close()

	// 等待进程退出
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}

	return nil
}
