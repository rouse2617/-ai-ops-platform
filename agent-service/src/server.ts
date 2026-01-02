import express, { Request, Response } from 'express';
import cors from 'cors';
import { Agent } from './agent/index.js';
import type { ChatRequest } from './types/index.js';

export function createServer(agent: Agent) {
  const app = express();

  app.use(cors());
  app.use(express.json());

  app.get('/health', (_req: Request, res: Response) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
  });

  app.post('/chat', async (req: Request, res: Response) => {
    try {
      const request: ChatRequest = req.body;
      const response = await agent.chat(request);
      res.json(response);
    } catch (error) {
      console.error('Chat error:', error);
      res.status(500).json({ error: 'Internal server error' });
    }
  });

  app.post('/chat/stream', async (req: Request, res: Response) => {
    try {
      const request: ChatRequest = req.body;

      res.setHeader('Content-Type', 'text/event-stream');
      res.setHeader('Cache-Control', 'no-cache');
      res.setHeader('Connection', 'keep-alive');

      for await (const chunk of agent.chatStream(request)) {
        res.write(`data: ${JSON.stringify(chunk)}\n\n`);
      }

      res.end();
    } catch (error) {
      console.error('Stream error:', error);
      res.write(`data: ${JSON.stringify({ type: 'error', error: 'Internal server error' })}\n\n`);
      res.end();
    }
  });

  return app;
}
