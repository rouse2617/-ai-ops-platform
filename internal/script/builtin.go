package script

import (
	"ai-ops/internal/model"
	"time"

	"github.com/google/uuid"
)

// GetBuiltinScripts 获取内置脚本列表
func GetBuiltinScripts() []*model.Script {
	now := time.Now()
	return []*model.Script{
		// 1. 系统信息收集
		{
			ID:          uuid.New().String(),
			Name:        "system_info",
			Description: "收集系统基本信息，包括主机名、操作系统、内核版本、CPU、内存、磁盘等",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
echo "=== 主机信息 ==="
echo "主机名: $(hostname)"
echo "操作系统: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'"' -f2 || uname -s)"
echo "内核版本: $(uname -r)"
echo "系统架构: $(uname -m)"
echo "运行时间: $(uptime -p 2>/dev/null || uptime)"
echo ""
echo "=== CPU 信息 ==="
echo "CPU 型号: $(grep 'model name' /proc/cpuinfo | head -1 | cut -d':' -f2 | xargs)"
echo "CPU 核心数: $(nproc)"
echo "CPU 使用率: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')%"
echo ""
echo "=== 内存信息 ==="
free -h
echo ""
echo "=== 磁盘信息 ==="
df -h | grep -v tmpfs | grep -v loop
echo ""
echo "=== 网络接口 ==="
ip -br addr 2>/dev/null || ifconfig -a 2>/dev/null | grep -E "^[a-z]|inet "`,
			Parameters: []model.ScriptParameter{},
			Enabled:    true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},

		// 2. 磁盘使用分析
		{
			ID:          uuid.New().String(),
			Name:        "disk_analyzer",
			Description: "分析磁盘使用情况，找出占用空间最大的目录和文件",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
TARGET_DIR="${1:-/}"
TOP_N="${2:-10}"

echo "=== 磁盘使用概览 ==="
df -h "$TARGET_DIR" 2>/dev/null || df -h /

echo ""
echo "=== 占用空间最大的 $TOP_N 个目录 ==="
du -h --max-depth=1 "$TARGET_DIR" 2>/dev/null | sort -hr | head -n "$TOP_N"

echo ""
echo "=== 占用空间最大的 $TOP_N 个文件 ==="
find "$TARGET_DIR" -type f -exec du -h {} + 2>/dev/null | sort -hr | head -n "$TOP_N"

echo ""
echo "=== 可清理的空间 ==="
echo "日志文件 (/var/log): $(du -sh /var/log 2>/dev/null | cut -f1)"
echo "临时文件 (/tmp): $(du -sh /tmp 2>/dev/null | cut -f1)"
echo "包缓存: $(du -sh /var/cache/apt 2>/dev/null | cut -f1 || du -sh /var/cache/yum 2>/dev/null | cut -f1)"`,
			Parameters: []model.ScriptParameter{
				{Name: "target_dir", Type: "string", Description: "要分析的目录路径", Required: false, Default: "/"},
				{Name: "top_n", Type: "integer", Description: "显示前 N 个结果", Required: false, Default: 10},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 3. 日志清理
		{
			ID:          uuid.New().String(),
			Name:        "log_cleaner",
			Description: "清理指定天数之前的日志文件，释放磁盘空间",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
DAYS="${1:-30}"
DRY_RUN="${2:-true}"

echo "=== 日志清理工具 ==="
echo "清理 $DAYS 天前的日志文件"
echo "模拟运行: $DRY_RUN"
echo ""

# 查找要清理的文件
echo "=== 将要清理的文件 ==="
find /var/log -type f -name "*.log" -mtime +$DAYS 2>/dev/null
find /var/log -type f -name "*.log.*" -mtime +$DAYS 2>/dev/null
find /var/log -type f -name "*.gz" -mtime +$DAYS 2>/dev/null

# 计算可释放空间
SPACE=$(find /var/log -type f \( -name "*.log" -o -name "*.log.*" -o -name "*.gz" \) -mtime +$DAYS -exec du -ch {} + 2>/dev/null | tail -1 | cut -f1)
echo ""
echo "可释放空间: $SPACE"

if [ "$DRY_RUN" = "false" ]; then
    echo ""
    echo "=== 开始清理 ==="
    find /var/log -type f -name "*.log" -mtime +$DAYS -delete 2>/dev/null
    find /var/log -type f -name "*.log.*" -mtime +$DAYS -delete 2>/dev/null
    find /var/log -type f -name "*.gz" -mtime +$DAYS -delete 2>/dev/null
    echo "清理完成！"
fi`,
			Parameters: []model.ScriptParameter{
				{Name: "days", Type: "integer", Description: "清理多少天前的日志", Required: false, Default: 30},
				{Name: "dry_run", Type: "boolean", Description: "是否模拟运行（不实际删除）", Required: false, Default: true},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 4. 进程监控
		{
			ID:          uuid.New().String(),
			Name:        "process_monitor",
			Description: "监控系统进程，显示 CPU 和内存占用最高的进程",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
TOP_N="${1:-10}"

echo "=== 系统负载 ==="
uptime

echo ""
echo "=== CPU 占用 TOP $TOP_N ==="
ps aux --sort=-%cpu | head -n $((TOP_N + 1)) | awk '{printf "%-10s %-8s %-6s %-6s %s\n", $1, $2, $3"%", $4"%", $11}'

echo ""
echo "=== 内存占用 TOP $TOP_N ==="
ps aux --sort=-%mem | head -n $((TOP_N + 1)) | awk '{printf "%-10s %-8s %-6s %-6s %s\n", $1, $2, $3"%", $4"%", $11}'

echo ""
echo "=== 僵尸进程 ==="
ZOMBIE=$(ps aux | awk '$8=="Z" {print $0}')
if [ -z "$ZOMBIE" ]; then
    echo "无僵尸进程"
else
    echo "$ZOMBIE"
fi

echo ""
echo "=== 进程统计 ==="
echo "总进程数: $(ps aux | wc -l)"
echo "运行中: $(ps aux | awk '$8=="R" {count++} END {print count+0}')"
echo "睡眠中: $(ps aux | awk '$8=="S" {count++} END {print count+0}')"`,
			Parameters: []model.ScriptParameter{
				{Name: "top_n", Type: "integer", Description: "显示前 N 个进程", Required: false, Default: 10},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 5. 网络诊断
		{
			ID:          uuid.New().String(),
			Name:        "network_diagnose",
			Description: "网络诊断工具，检查网络连接、端口监听、DNS 解析等",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
TARGET_HOST="${1:-8.8.8.8}"

echo "=== 网络接口状态 ==="
ip -br link 2>/dev/null || ifconfig | grep -E "^[a-z]"

echo ""
echo "=== IP 地址 ==="
ip -br addr 2>/dev/null || ifconfig | grep "inet "

echo ""
echo "=== 路由表 ==="
ip route 2>/dev/null || netstat -rn

echo ""
echo "=== DNS 配置 ==="
cat /etc/resolv.conf | grep -v "^#"

echo ""
echo "=== 连通性测试 (ping $TARGET_HOST) ==="
ping -c 3 "$TARGET_HOST" 2>/dev/null || echo "ping 失败"

echo ""
echo "=== 监听端口 ==="
ss -tlnp 2>/dev/null || netstat -tlnp 2>/dev/null

echo ""
echo "=== 活动连接统计 ==="
ss -s 2>/dev/null || netstat -s | head -20

echo ""
echo "=== 连接状态分布 ==="
ss -tan 2>/dev/null | awk 'NR>1 {print $1}' | sort | uniq -c | sort -rn`,
			Parameters: []model.ScriptParameter{
				{Name: "target_host", Type: "string", Description: "测试连通性的目标主机", Required: false, Default: "8.8.8.8"},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 6. 服务状态检查
		{
			ID:          uuid.New().String(),
			Name:        "service_checker",
			Description: "检查常见服务的运行状态",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
SERVICES="${1:-nginx,mysql,redis,docker,sshd}"

echo "=== 服务状态检查 ==="
echo ""

IFS=',' read -ra SERVICE_LIST <<< "$SERVICES"
for service in "${SERVICE_LIST[@]}"; do
    service=$(echo "$service" | xargs)  # 去除空格
    if systemctl is-active --quiet "$service" 2>/dev/null; then
        STATUS="✓ 运行中"
    elif systemctl list-unit-files | grep -q "^$service"; then
        STATUS="✗ 已停止"
    else
        STATUS="- 未安装"
    fi
    printf "%-20s %s\n" "$service" "$STATUS"
done

echo ""
echo "=== 失败的服务 ==="
systemctl --failed 2>/dev/null | head -20

echo ""
echo "=== 最近启动的服务 ==="
systemctl list-units --type=service --state=running 2>/dev/null | head -15`,
			Parameters: []model.ScriptParameter{
				{Name: "services", Type: "string", Description: "要检查的服务列表，逗号分隔", Required: false, Default: "nginx,mysql,redis,docker,sshd"},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 7. 安全检查
		{
			ID:          uuid.New().String(),
			Name:        "security_audit",
			Description: "基础安全检查，包括用户、权限、SSH 配置等",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
echo "=== 安全审计报告 ==="
echo "检查时间: $(date)"
echo ""

echo "=== 系统用户检查 ==="
echo "可登录用户:"
grep -v '/nologin\|/false' /etc/passwd | cut -d: -f1,7

echo ""
echo "空密码用户:"
awk -F: '($2 == "" || $2 == "!") {print $1}' /etc/shadow 2>/dev/null || echo "无权限检查"

echo ""
echo "=== SSH 配置检查 ==="
if [ -f /etc/ssh/sshd_config ]; then
    echo "Root 登录: $(grep -i "^PermitRootLogin" /etc/ssh/sshd_config || echo "默认")"
    echo "密码认证: $(grep -i "^PasswordAuthentication" /etc/ssh/sshd_config || echo "默认")"
    echo "SSH 端口: $(grep -i "^Port" /etc/ssh/sshd_config || echo "22 (默认)")"
fi

echo ""
echo "=== SUID 文件 ==="
find /usr -perm -4000 -type f 2>/dev/null | head -10
echo "(仅显示前10个)"

echo ""
echo "=== 最近登录 ==="
last -n 10 2>/dev/null

echo ""
echo "=== 失败登录尝试 ==="
lastb -n 10 2>/dev/null || grep "Failed password" /var/log/auth.log 2>/dev/null | tail -10

echo ""
echo "=== 开放端口 ==="
ss -tlnp 2>/dev/null | grep LISTEN`,
			Parameters: []model.ScriptParameter{},
			Enabled:    true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},

		// 8. 配置备份
		{
			ID:          uuid.New().String(),
			Name:        "config_backup",
			Description: "备份系统重要配置文件到指定目录",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
BACKUP_DIR="${1:-/tmp/config_backup}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="$BACKUP_DIR/backup_$DATE"

echo "=== 配置文件备份 ==="
echo "备份目录: $BACKUP_PATH"
echo ""

mkdir -p "$BACKUP_PATH"

# 备份配置文件
CONFIG_FILES=(
    "/etc/passwd"
    "/etc/group"
    "/etc/hosts"
    "/etc/fstab"
    "/etc/crontab"
    "/etc/ssh/sshd_config"
    "/etc/nginx/nginx.conf"
    "/etc/my.cnf"
    "/etc/redis/redis.conf"
)

echo "=== 备份文件列表 ==="
for file in "${CONFIG_FILES[@]}"; do
    if [ -f "$file" ]; then
        cp "$file" "$BACKUP_PATH/" 2>/dev/null && echo "✓ $file"
    fi
done

# 备份 crontab
crontab -l > "$BACKUP_PATH/crontab_$(whoami)" 2>/dev/null && echo "✓ crontab"

# 备份 iptables
iptables-save > "$BACKUP_PATH/iptables.rules" 2>/dev/null && echo "✓ iptables"

echo ""
echo "=== 备份完成 ==="
echo "备份大小: $(du -sh "$BACKUP_PATH" | cut -f1)"
ls -la "$BACKUP_PATH"`,
			Parameters: []model.ScriptParameter{
				{Name: "backup_dir", Type: "string", Description: "备份存储目录", Required: false, Default: "/tmp/config_backup"},
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// 9. Docker 状态检查
		{
			ID:          uuid.New().String(),
			Name:        "docker_status",
			Description: "检查 Docker 容器和镜像状态",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
echo "=== Docker 状态检查 ==="

if ! command -v docker &> /dev/null; then
    echo "Docker 未安装"
    exit 1
fi

echo ""
echo "=== Docker 版本 ==="
docker version --format '{{.Server.Version}}' 2>/dev/null || echo "无法获取版本"

echo ""
echo "=== 运行中的容器 ==="
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null

echo ""
echo "=== 所有容器 ==="
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Image}}" 2>/dev/null

echo ""
echo "=== 镜像列表 ==="
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" 2>/dev/null

echo ""
echo "=== 磁盘使用 ==="
docker system df 2>/dev/null

echo ""
echo "=== 容器资源使用 ==="
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null`,
			Parameters: []model.ScriptParameter{},
			Enabled:    true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},

		// 10. 性能基准测试
		{
			ID:          uuid.New().String(),
			Name:        "performance_benchmark",
			Description: "简单的系统性能基准测试",
			Language:    model.ScriptLanguageBash,
			Content: `#!/bin/bash
echo "=== 系统性能基准测试 ==="
echo "测试时间: $(date)"
echo ""

echo "=== CPU 性能测试 ==="
echo "计算 10000 以内的质数..."
START=$(date +%s.%N)
seq 2 10000 | factor | awk 'NF==2' | wc -l
END=$(date +%s.%N)
echo "耗时: $(echo "$END - $START" | bc) 秒"

echo ""
echo "=== 内存带宽测试 ==="
if command -v dd &> /dev/null; then
    echo "写入 100MB 到内存..."
    dd if=/dev/zero of=/dev/null bs=1M count=100 2>&1 | tail -1
fi

echo ""
echo "=== 磁盘 I/O 测试 ==="
TEST_FILE="/tmp/disk_test_$$"
echo "写入测试 (100MB)..."
dd if=/dev/zero of="$TEST_FILE" bs=1M count=100 conv=fdatasync 2>&1 | tail -1
echo "读取测试..."
dd if="$TEST_FILE" of=/dev/null bs=1M 2>&1 | tail -1
rm -f "$TEST_FILE"

echo ""
echo "=== 网络延迟测试 ==="
for host in "8.8.8.8" "114.114.114.114"; do
    echo -n "$host: "
    ping -c 3 -q "$host" 2>/dev/null | tail -1 | awk -F'/' '{print $5 " ms"}' || echo "不可达"
done`,
			Parameters: []model.ScriptParameter{},
			Enabled:    true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
}
