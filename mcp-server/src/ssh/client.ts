interface SSHExecuteRequest {
  host: string;
  command: string;
}

interface GoBackendResponse {
  code: number;
  message: string;
  data: {
    host: string;
    command: string;
    output: string;
  };
}

interface SSHExecuteResponse {
  output: string;
  error?: string;
}

export class SSHClient {
  private readonly backendUrl: string;

  constructor(backendUrl = 'http://localhost:1280') {
    this.backendUrl = backendUrl;
  }

  async execute(host: string, command: string): Promise<SSHExecuteResponse> {
    const response = await fetch(`${this.backendUrl}/api/internal/ssh/execute`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ host, command } satisfies SSHExecuteRequest)
    });

    if (!response.ok) {
      throw new Error(`SSH execution failed: ${response.statusText}`);
    }

    const result = await response.json() as GoBackendResponse;

    if (result.code !== 0) {
      return { output: '', error: result.message };
    }

    return { output: result.data.output };
  }
}
