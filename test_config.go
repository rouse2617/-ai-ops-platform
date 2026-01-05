package main

import (
	"fmt"
	"ai-ops/internal/config"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	fmt.Printf("MCP 配置数量: %d\n", len(cfg.MCP))
	for i, mcp := range cfg.MCP {
		fmt.Printf("MCP[%d]: Name=%s, URL=%s, Timeout=%d, Enabled=%v\n",
			i, mcp.Name, mcp.URL, mcp.Timeout, mcp.Enabled)
	}
}
