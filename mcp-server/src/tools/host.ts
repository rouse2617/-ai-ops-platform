interface Host {
  id: number;
  name: string;
  ip: string;
  status: string;
}

export class HostTools {
  private readonly backendUrl: string;

  constructor(backendUrl = 'http://localhost:8080') {
    this.backendUrl = backendUrl;
  }

  async listHosts(): Promise<string> {
    const response = await fetch(`${this.backendUrl}/api/hosts`);

    if (!response.ok) {
      throw new Error(`Failed to fetch hosts: ${response.statusText}`);
    }

    const hosts = await response.json() as Host[];
    return JSON.stringify(hosts, null, 2);
  }
}
