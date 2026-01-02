# AI-Ops MCP Server

MCP Server providing operations tools for the AI-Ops platform.

## Features

- System monitoring (CPU, memory, disk)
- Log querying
- Process management
- Command execution
- Host management

## Installation

```bash
npm install
npm run build
```

## Usage

```bash
npm start
```

## Environment Variables

- `PORT`: Server port (default: 3001)
- `BACKEND_URL`: Go backend URL (default: http://localhost:8080)

## Docker

```bash
docker build -t ai-ops-mcp-server .
docker run -p 3001:3001 -e BACKEND_URL=http://backend:8080 ai-ops-mcp-server
```

## Tools

- `check_cpu`: Check CPU usage
- `check_memory`: Check memory usage
- `check_disk`: Check disk usage
- `query_log`: Query log files
- `check_process`: Check process status
- `run_command`: Execute commands
- `list_hosts`: List all hosts
