// 危险命令检测工具

// 危险命令关键词列表
const DANGEROUS_PATTERNS = [
  // 删除命令
  /\brm\s+(-rf|-r\s+-f|-fr|-f\s+-r)\b/i,
  /\brm\s+-r\b/i,
  /\brm\s+.*\/\s*$/i, // rm ... /

  // 重启/关机命令
  /\breboot\b/i,
  /\bshutdown\b/i,
  /\bhalt\b/i,
  /\bpoweroff\b/i,
  /\binit\s+[06]\b/i,

  // 服务停止命令
  /\bsystemctl\s+stop\b/i,
  /\bservice\s+\S+\s+stop\b/i,
  /\bkillall\b/i,
  /\bkill\s+-9\b/i,
  /\bpkill\s+-9\b/i,

  // 格式化/文件系统操作
  /\bmkfs\b/i,
  /\bfdisk\b/i,
  /\bdd\s+if=/i,

  // 网络操作
  /\broute\s+del\b/i,
  /\bip\s+route\s+del\b/i,

  // 权限修改
  /\bchmod\s+000\b/i,
  /\bchmod\s+-R\s+000\b/i,

  // 数据库操作
  /\bdrop\s+database\b/i,
  /\btruncate\s+table\b/i,
  /\bdelete\s+from\s+.*\s+where\s+1\s*=\s*1/i,
]

// 高危险命令（需要强制确认）
const CRITICAL_PATTERNS = [
  /\brm\s+(-rf|-r\s+-f|-fr|-f\s+-r)\b/i,
  /\breboot\b/i,
  /\bshutdown\b/i,
  /\bhalt\b/i,
  /\bpoweroff\b/i,
  /\bmkfs\b/i,
  /\bdd\s+if=/i,
]

/**
 * 检测命令是否危险
 * @param command 命令字符串
 * @returns 是否危险
 */
export function isDangerousCommand(command: string): boolean {
  if (!command || !command.trim()) {
    return false
  }
  
  const trimmed = command.trim()
  return DANGEROUS_PATTERNS.some(pattern => pattern.test(trimmed))
}

/**
 * 检测命令是否是高危险命令（需要强制确认）
 * @param command 命令字符串
 * @returns 是否高危险
 */
export function isCriticalCommand(command: string): boolean {
  if (!command || !command.trim()) {
    return false
  }
  
  const trimmed = command.trim()
  return CRITICAL_PATTERNS.some(pattern => pattern.test(trimmed))
}

/**
 * 获取危险命令的警告信息
 * @param command 命令字符串
 * @returns 警告信息
 */
export function getDangerousCommandWarning(command: string): string {
  if (isCriticalCommand(command)) {
    return '这是一个高危险命令，可能会造成不可逆的数据丢失或系统故障！'
  }
  
  if (isDangerousCommand(command)) {
    return '这是一个危险命令，请确认操作是否正确！'
  }
  
  return ''
}

/**
 * 高亮危险关键词（返回HTML字符串）
 * @param command 命令字符串
 * @returns 高亮后的HTML字符串
 */
export function highlightDangerousKeywords(command: string): string {
  if (!command) {
    return ''
  }

  let highlighted = command
  const dangerKeywords = ['rm -rf', 'reboot', 'shutdown', 'kill -9', 'mkfs', 'dd if=']

  dangerKeywords.forEach(keyword => {
    const regex = new RegExp(`(${keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    highlighted = highlighted.replace(regex, '<span style="color: #f56c6c; font-weight: bold;">$1</span>')
  })

  return highlighted
}

/**
 * 获取命令风险等级
 * @param command 命令字符串
 * @returns 风险等级
 */
export function getCommandRiskLevel(command: string): 'low' | 'medium' | 'high' | 'critical' {
  if (!command || !command.trim()) {
    return 'low'
  }

  const trimmed = command.trim()

  // 极高风险命令
  if (/\brm\s+(-rf|-r\s+-f|-fr)\s+\/\s*$/i.test(trimmed) ||
      /\bmkfs\s+/i.test(trimmed) ||
      /\bdd\s+if=.*of=\/dev\//i.test(trimmed)) {
    return 'critical'
  }

  // 高风险命令
  if (isCriticalCommand(trimmed)) {
    return 'high'
  }

  // 中等风险命令
  if (isDangerousCommand(trimmed)) {
    return 'medium'
  }

  return 'low'
}

/**
 * 获取命令影响范围描述
 * @param command 命令字符串
 * @returns 影响范围描述
 */
export function getCommandImpact(command: string): string {
  if (!command || !command.trim()) {
    return ''
  }

  const trimmed = command.trim()

  // 删除命令
  if (/\brm\s+(-rf|-r\s+-f|-fr)\s+\/\s*$/i.test(trimmed)) {
    return '将删除整个根文件系统，系统将完全不可用'
  }
  if (/\brm\s+(-rf|-r\s+-f|-fr)\s+/i.test(trimmed)) {
    return '将递归删除指定目录及其所有内容，数据不可恢复'
  }

  // 重启/关机命令
  if (/\breboot\b/i.test(trimmed)) {
    return '将立即重启系统，所有运行中的服务将中断'
  }
  if (/\bshutdown\b/i.test(trimmed) || /\bhalt\b/i.test(trimmed) || /\bpoweroff\b/i.test(trimmed)) {
    return '将关闭系统，所有服务将停止'
  }

  // 服务停止
  if (/\bsystemctl\s+stop\s+/i.test(trimmed)) {
    return '将停止指定的系统服务，可能影响业务运行'
  }

  // 进程终止
  if (/\bkillall\s+/i.test(trimmed) || /\bkill\s+-9\s+/i.test(trimmed)) {
    return '将强制终止进程，可能导致数据丢失'
  }

  // 格式化
  if (/\bmkfs\s+/i.test(trimmed)) {
    return '将格式化磁盘分区，所有数据将被清除'
  }

  // dd 命令
  if (/\bdd\s+if=/i.test(trimmed)) {
    return '将进行底层磁盘操作，可能覆盖重要数据'
  }

  // 数据库操作
  if (/\bdrop\s+database\s+/i.test(trimmed)) {
    return '将删除整个数据库，所有数据将丢失'
  }
  if (/\btruncate\s+table\s+/i.test(trimmed)) {
    return '将清空数据表，所有记录将被删除'
  }

  return '可能对系统造成影响，请谨慎操作'
}













