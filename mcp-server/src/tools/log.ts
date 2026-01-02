import { DirectSSHClient } from '../ssh/direct-client.js';

export interface LogToolArgs {
  host: string;
  logPath: string;
  keyword?: string;
  lines?: number;
}

export class LogTools {
  constructor(private ssh: DirectSSHClient) {}

  async queryLog(args: LogToolArgs): Promise<string> {
    const lines = args.lines || 100;
    let command = `tail -n ${lines} ${args.logPath}`;

    if (args.keyword) {
      command += ` | grep '${args.keyword}'`;
    }

    const result = await this.ssh.execute(args.host, command);
    return result.error || result.output;
  }
}
