import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { ListToolsRequestSchema, CallToolRequestSchema } from '@modelcontextprotocol/sdk/types.js';

// 工具列表定义
const toolsList = [
  {
    name: 'check_cpu',
    description: 'Check CPU usage on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' }
      },
      required: ['host']
    }
  },
  {
    name: 'check_memory',
    description: 'Check memory usage on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' }
      },
      required: ['host']
    }
  },
  {
    name: 'check_disk',
    description: 'Check disk usage on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' }
      },
      required: ['host']
    }
  },
  {
    name: 'query_log',
    description: 'Query log files on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' },
        logPath: { type: 'string', description: 'Log file path' },
        keyword: { type: 'string', description: 'Search keyword (optional)' },
        lines: { type: 'number', description: 'Number of lines (default: 100)' }
      },
      required: ['host', 'logPath']
    }
  },
  {
    name: 'check_process',
    description: 'Check if a process is running on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' },
        processName: { type: 'string', description: 'Process name' }
      },
      required: ['host', 'processName']
    }
  },
  {
    name: 'run_command',
    description: 'Execute a command on a host',
    input_schema: {
      type: 'object' as const,
      properties: {
        host: { type: 'string', description: 'Host name or IP' },
        command: { type: 'string', description: 'Command to execute' }
      },
      required: ['host', 'command']
    }
  },
  {
    name: 'list_hosts',
    description: 'List all available hosts',
    input_schema: {
      type: 'object' as const,
      properties: {}
    }
  }
];

// 导出工具列表
export function getToolsList() {
  return toolsList;
}

export function registerTools(server: Server, backendUrl: string): void {
  server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: toolsList.map(t => ({
      name: t.name,
      description: t.description,
      inputSchema: t.input_schema
    }))
  }));

  server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const { name, arguments: args } = request.params;

    try {
      const { executeToolByName } = await import('../index.js');
      const result = await executeToolByName(name, args || {}, backendUrl);

      return {
        content: [{ type: 'text', text: typeof result === 'string' ? result : JSON.stringify(result) }]
      };
    } catch (error) {
      return {
        content: [{ type: 'text', text: `Error: ${error instanceof Error ? error.message : String(error)}` }],
        isError: true
      };
    }
  });
}
