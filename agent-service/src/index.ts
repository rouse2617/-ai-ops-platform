import dotenv from 'dotenv';
import { Agent } from './agent/index.js';
import { createServer } from './server.js';

dotenv.config();

async function main() {
  const agent = new Agent();
  await agent.initialize();

  const app = createServer(agent);
  const port = process.env.PORT || 3001;

  app.listen(port, () => {
    console.log(`Agent service running on port ${port}`);
  });
}

main().catch(console.error);
