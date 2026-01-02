# Agent Service

Node.js Agent Service using Claude Agent SDK for AI-Ops project.

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run development server:
```bash
npm run dev
```

4. Build for production:
```bash
npm run build
npm start
```

## API Endpoints

- `POST /chat` - Non-streaming chat
- `POST /chat/stream` - Streaming chat (SSE)
- `GET /health` - Health check

## Docker

```bash
docker build -t agent-service .
docker run -p 3001:3001 --env-file .env agent-service
```
