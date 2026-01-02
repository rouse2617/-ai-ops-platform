import { DirectSSHClient } from '../ssh/direct-client.js';

export interface SystemToolArgs {
  host: string;
}

export class SystemTools {
  constructor(private ssh: DirectSSHClient) {}

  async checkCpu(args: SystemToolArgs): Promise<string> {
    const result = await this.ssh.execute(
      args.host,
      "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"
    );
    return result.error || `CPU Usage: ${result.output.trim()}%`;
  }

  async checkMemory(args: SystemToolArgs): Promise<string> {
    const result = await this.ssh.execute(
      args.host,
      "free -m | awk 'NR==2{printf \"Used: %sMB (%.2f%%)\\n\", $3, $3*100/$2}'"
    );
    return result.error || result.output.trim();
  }

  async checkDisk(args: SystemToolArgs): Promise<string> {
    const result = await this.ssh.execute(
      args.host,
      "df -h | awk '$NF==\"/\"{printf \"Disk: %s/%s (%s)\\n\", $3, $2, $5}'"
    );
    return result.error || result.output.trim();
  }
}
