import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { registerTools, getToolsList } from './tools/index.js';

const BACKEND_URL = process.env.GO_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:1280';

const server = new Server(
  {
    name: 'ai-ops-mcp-server',
    version: '1.0.0'
  },
  {
    capabilities: {
      tools: {}
    }
  }
);

registerTools(server, BACKEND_URL);

const transport = new StdioServerTransport();
server.connect(transport).catch(console.error);

// Instance cache
const instanceCache = new Map<string, {
  ssh: any;
  systemTools: any;
  logTools: any;
  processTools: any;
  hostTools: any;
}>();

async function getInstances(backendUrl: string) {
  if (instanceCache.has(backendUrl)) {
    return instanceCache.get(backendUrl)!;
  }

  const { DirectSSHClient } = await import('./ssh/direct-client.js');
  const { SystemTools } = await import('./tools/system.js');
  const { LogTools } = await import('./tools/log.js');
  const { ProcessTools } = await import('./tools/process.js');
  const { HostTools } = await import('./tools/host.js');

  const ssh = new DirectSSHClient(backendUrl);
  const instances = {
    ssh,
    systemTools: new SystemTools(ssh),
    logTools: new LogTools(ssh),
    processTools: new ProcessTools(ssh),
    hostTools: new HostTools(backendUrl)
  };

  instanceCache.set(backendUrl, instances);
  return instances;
}

export async function executeToolByName(name: string, input: Record<string, unknown>, backendUrl: string): Promise<unknown> {
  const { systemTools, logTools, processTools, hostTools } = await getInstances(backendUrl);

  switch (name) {
    case 'check_cpu':
      return { result: await systemTools.checkCpu(input as { host: string }) };
    case 'check_memory':
      return { result: await systemTools.checkMemory(input as { host: string }) };
    case 'check_disk':
      return { result: await systemTools.checkDisk(input as { host: string }) };
    case 'query_log':
      return { result: await logTools.queryLog(input as any) };
    case 'check_process':
      return { result: await processTools.checkProcess(input as any) };
    case 'run_command':
      return { result: await processTools.runCommand(input as any) };
    case 'list_hosts':
      return { result: await hostTools.listHosts() };
    default:
      throw new Error(`Unknown tool: ${name}`);
  }
}
