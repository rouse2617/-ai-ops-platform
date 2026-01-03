// Cypress configuration and environment setup

export const TEST_CONFIG = {
  baseUrl: 'http://localhost:1281',
  apiUrl: 'http://localhost:1280/api',
  timeout: 10000,
  hosts: {
    'host-001': { name: 'Server 1', status: 'online' },
    'host-002': { name: 'Server 2', status: 'offline' },
    'host-003': { name: 'Server 3', status: 'online' }
  }
}

export const DANGEROUS_COMMANDS = [
  'rm -rf /tmp/test',
  'mkfs.ext4 /dev/sdb1',
  'reboot',
  'shutdown -h now',
  'kill -9 1'
]

export const SAFE_COMMANDS = [
  '查看系统负载',
  '查看进程列表',
  '查看网络状态',
  '分析系统日志'
]

export const LONG_OUTPUT = Array(100).fill(0).map((_, i) => `Line ${i + 1}: Sample output line`).join('\n')

export const SHORT_OUTPUT = Array(20).fill(0).map((_, i) => `Line ${i + 1}: Sample output line`).join('\n')
