import Anthropic from '@anthropic-ai/sdk';
import type { MessageParam, Tool, ToolUseBlock, ToolResultBlockParam } from '@anthropic-ai/sdk/resources/messages';
import type { ChatRequest, ChatResponse, StreamChunk, MCPTool, ToolCall } from '../types/index.js';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export class Agent {
  private client: Anthropic;
  private model: string;
  private goBackendUrl: string;
  private tools: Tool[] = [];
  private lastToolsRefresh: number = 0;
  private readonly REFRESH_INTERVAL = 5 * 60 * 1000; // 5 minutes
  private mcpClient!: Client;

  constructor() {
    this.client = new Anthropic({
      apiKey: process.env.ANTHROPIC_API_KEY,
      baseURL: process.env.ANTHROPIC_BASE_URL,
    });
    this.model = process.env.LLM_MODEL || 'claude-sonnet-4-5-20250929';
    this.goBackendUrl = process.env.GO_BACKEND_URL || 'http://localhost:1280';
  }

  async initialize(): Promise<void> {
    await this.initializeMCPClient();
    await this.loadMCPTools();
  }

  private async initializeMCPClient(): Promise<void> {
    const mcpServerPath = path.resolve(__dirname, '../../../mcp-server/dist/index.js');
    const transport = new StdioClientTransport({
      command: 'node',
      args: [mcpServerPath],
      env: {
        ...process.env,
        MCP_MODE: 'stdio',
        GO_BACKEND_URL: this.goBackendUrl,
      },
    });
    this.mcpClient = new Client({ name: 'agent-service', version: '1.0.0' }, { capabilities: {} });
    await this.mcpClient.connect(transport);
  }

  private async loadMCPTools(): Promise<void> {
    try {
      const result = await this.mcpClient.listTools();
      this.tools = result.tools.map(tool => ({
        name: tool.name,
        description: tool.description,
        input_schema: tool.inputSchema,
      }));
      this.lastToolsRefresh = Date.now();
    } catch (error) {
      console.error('Failed to load MCP tools:', error);
    }
  }

  private async refreshToolsIfNeeded(): Promise<void> {
    if (Date.now() - this.lastToolsRefresh > this.REFRESH_INTERVAL) {
      await this.loadMCPTools();
    }
  }

  private async callTool(name: string, input: Record<string, unknown>): Promise<unknown> {
    const result = await this.mcpClient.callTool({ name, arguments: input });
    return result.content;
  }

  private buildMessages(request: ChatRequest): MessageParam[] {
    const messages: MessageParam[] = [];

    if (request.history) {
      messages.push(...request.history.map(msg => ({
        role: msg.role,
        content: msg.content,
      })));
    }

    let userMessage = request.message;
    if (request.hosts && request.hosts.length > 0) {
      userMessage = `[Context: hosts=${request.hosts.join(',')}]\n${request.message}`;
    }

    messages.push({ role: 'user', content: userMessage });
    return messages;
  }

  async chat(request: ChatRequest): Promise<ChatResponse> {
    await this.refreshToolsIfNeeded();
    const messages = this.buildMessages(request);
    const toolCalls: ToolCall[] = [];
    let finalResponse = '';

    let currentMessages = [...messages];

    while (true) {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 4096,
        messages: currentMessages,
        tools: this.tools.length > 0 ? this.tools : undefined,
      });

      if (response.stop_reason === 'end_turn') {
        const textContent = response.content.find(c => c.type === 'text');
        if (textContent && textContent.type === 'text') {
          finalResponse = textContent.text;
        }
        break;
      }

      if (response.stop_reason === 'tool_use') {
        const assistantMessage: MessageParam = {
          role: 'assistant',
          content: response.content,
        };
        currentMessages.push(assistantMessage);

        const toolResults = [];
        for (const block of response.content) {
          if (block.type === 'tool_use') {
            const result = await this.callTool(block.name, block.input as Record<string, unknown>);
            toolCalls.push({
              name: block.name,
              input: block.input as Record<string, unknown>,
              result,
            });
            toolResults.push({
              type: 'tool_result' as const,
              tool_use_id: block.id,
              content: JSON.stringify(result),
            });
          }
        }

        currentMessages.push({
          role: 'user',
          content: toolResults,
        });
      } else {
        break;
      }
    }

    return { response: finalResponse, toolCalls };
  }

  async *chatStream(request: ChatRequest): AsyncGenerator<StreamChunk> {
    await this.refreshToolsIfNeeded();
    const messages = this.buildMessages(request);
    let currentMessages = [...messages];

    while (true) {
      console.log('[chatStream] Starting new iteration, messages count:', currentMessages.length);
      const stream = await this.client.messages.create({
        model: this.model,
        max_tokens: 4096,
        messages: currentMessages,
        tools: this.tools.length > 0 ? this.tools : undefined,
        stream: true,
      });

      let currentToolUse: { id: string; name: string; input: string } | null = null;
      const assistantContent: Array<ToolUseBlock> = [];
      let stopReason: string | null = null;

      for await (const event of stream) {
        if (event.type === 'content_block_start') {
          if (event.content_block.type === 'tool_use') {
            currentToolUse = {
              id: event.content_block.id,
              name: event.content_block.name,
              input: '',
            };
          }
        } else if (event.type === 'content_block_delta') {
          if (event.delta.type === 'text_delta') {
            yield { type: 'text', content: event.delta.text };
          } else if (event.delta.type === 'input_json_delta' && currentToolUse) {
            currentToolUse.input += event.delta.partial_json;
          }
        } else if (event.type === 'content_block_stop' && currentToolUse) {
          const input = JSON.parse(currentToolUse.input || '{}');
          assistantContent.push({
            type: 'tool_use',
            id: currentToolUse.id,
            name: currentToolUse.name,
            input,
          });
          yield {
            type: 'tool_use',
            toolCall: { name: currentToolUse.name, input },
          };
          currentToolUse = null;
        } else if (event.type === 'message_delta') {
          stopReason = event.delta.stop_reason;
          console.log('[chatStream] Got stop_reason:', stopReason);
        }
      }

      console.log('[chatStream] Stream ended, stopReason:', stopReason, 'assistantContent:', assistantContent.length);

      // If we have tool calls, process them regardless of stop_reason
      // Some API proxies may return end_turn instead of tool_use
      if (assistantContent.length > 0) {
        console.log('[chatStream] Processing tool calls...');
        currentMessages.push({ role: 'assistant', content: assistantContent });

        const toolResults: ToolResultBlockParam[] = [];
        for (const block of assistantContent) {
          console.log('[chatStream] Calling tool:', block.name);
          const result = await this.callTool(block.name, block.input as Record<string, unknown>);
          console.log('[chatStream] Tool result:', result);
          toolResults.push({
            type: 'tool_result',
            tool_use_id: block.id,
            content: JSON.stringify(result),
          });
          yield { type: 'tool_result', toolCall: { name: block.name, input: block.input as Record<string, unknown>, result } };
        }

        currentMessages.push({ role: 'user', content: toolResults });
        console.log('[chatStream] Tool results added, continuing loop...');
        // Continue the loop to get Claude's response after tool use
        continue;
      }

      // No tool calls, we're done
      console.log('[chatStream] No tool calls, yielding done');
      yield { type: 'done' };
      return;
    }
  }
}
