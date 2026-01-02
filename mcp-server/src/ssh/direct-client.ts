import { Client, ConnectConfig } from 'ssh2';

interface HostCredentials {
  ip: string;
  port: number;
  username: string;
  password?: string;
  privateKey?: string;
}

interface SSHExecuteResponse {
  output: string;
  error?: string;
}

interface ConnectionCacheEntry {
  client: Client;
  lastUsed: number;
}

export class DirectSSHClient {
  private readonly backendUrl: string;
  private readonly connectionCache = new Map<string, ConnectionCacheEntry>();
  private readonly cacheTimeout = 5 * 60 * 1000; // 5 minutes

  constructor(backendUrl = 'http://localhost:1280') {
    this.backendUrl = backendUrl;
    this.startCleanupTimer();
  }

  async execute(host: string, command: string): Promise<SSHExecuteResponse> {
    try {
      const credentials = await this.getCredentials(host);
      const client = await this.getConnection(host, credentials);

      return await this.executeCommand(client, command);
    } catch (error) {
      return {
        output: '',
        error: error instanceof Error ? error.message : String(error)
      };
    }
  }

  private async getCredentials(host: string): Promise<HostCredentials> {
    const response = await fetch(`${this.backendUrl}/api/hosts/all`);

    if (!response.ok) {
      throw new Error(`Failed to fetch hosts: ${response.statusText}`);
    }

    const result = await response.json() as { code: number; data?: any[] };

    if (result.code !== 0) {
      throw new Error('Failed to get host list');
    }

    if (!result.data || !Array.isArray(result.data)) {
      throw new Error('Invalid host list format: data is not an array');
    }

    const hostData = result.data.find((h: any) => h.host === host);

    if (!hostData) {
      throw new Error(`Host ${host} not found`);
    }

    return {
      ip: hostData.host,
      port: hostData.port || 22,
      username: hostData.username || hostData.user,
      password: hostData.password,
      privateKey: hostData.privateKey || hostData.private_key
    };
  }

  private async getConnection(host: string, credentials: HostCredentials): Promise<Client> {
    const cached = this.connectionCache.get(host);

    if (cached && this.isConnectionAlive(cached.client)) {
      cached.lastUsed = Date.now();
      return cached.client;
    }

    const client = await this.createConnection(credentials);
    this.connectionCache.set(host, { client, lastUsed: Date.now() });

    return client;
  }

  private createConnection(credentials: HostCredentials): Promise<Client> {
    return new Promise((resolve, reject) => {
      const client = new Client();

      const config: ConnectConfig = {
        host: credentials.ip,
        port: credentials.port,
        username: credentials.username,
        readyTimeout: 10000
      };

      if (credentials.privateKey) {
        config.privateKey = credentials.privateKey;
      } else if (credentials.password) {
        config.password = credentials.password;
      }

      client.on('ready', () => resolve(client));
      client.on('error', reject);

      client.connect(config);
    });
  }

  private executeCommand(client: Client, command: string): Promise<SSHExecuteResponse> {
    return new Promise((resolve) => {
      client.exec(command, (err, stream) => {
        if (err) {
          resolve({ output: '', error: err.message });
          return;
        }

        let stdout = '';
        let stderr = '';

        stream.on('data', (data: Buffer) => {
          stdout += data.toString();
        });

        stream.stderr.on('data', (data: Buffer) => {
          stderr += data.toString();
        });

        stream.on('close', () => {
          if (stderr) {
            resolve({ output: stdout, error: stderr });
          } else {
            resolve({ output: stdout });
          }
        });
      });
    });
  }

  private isConnectionAlive(client: Client): boolean {
    try {
      return !!client;
    } catch {
      return false;
    }
  }

  private startCleanupTimer(): void {
    setInterval(() => {
      const now = Date.now();

      for (const [host, entry] of this.connectionCache.entries()) {
        if (now - entry.lastUsed > this.cacheTimeout) {
          entry.client.end();
          this.connectionCache.delete(host);
        }
      }
    }, 60000); // Check every minute
  }

  destroy(): void {
    for (const entry of this.connectionCache.values()) {
      entry.client.end();
    }
    this.connectionCache.clear();
  }
}
