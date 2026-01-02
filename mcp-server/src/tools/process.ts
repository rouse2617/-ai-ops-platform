import { DirectSSHClient } from '../ssh/direct-client.js';

export interface ProcessCheckArgs {
  host: string;
  processName: string;
}

export interface RunCommandArgs {
  host: string;
  command: string;
}

export class ProcessTools {
  constructor(private ssh: DirectSSHClient) {}

  async checkProcess(args: ProcessCheckArgs): Promise<string> {
    const result = await this.ssh.execute(
      args.host,
      `ps aux | grep '${args.processName}' | grep -v grep`
    );
    return result.error || (result.output.trim() || 'Process not found');
  }

  async runCommand(args: RunCommandArgs): Promise<string> {
    const result = await this.ssh.execute(args.host, args.command);
    return result.error || result.output;
  }
}
