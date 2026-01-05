package builtin

import (
	"fmt"
	"strings"

	"ai-ops/internal/tool"
)

// CheckMySQLStatusTool MySQL 状态检查工具
type CheckMySQLStatusTool struct{}

func (t *CheckMySQLStatusTool) Name() string { return "check_mysql_status" }

func (t *CheckMySQLStatusTool) Description() string {
	return `检查 MySQL 数据库状态。
返回 MySQL 运行状态、连接数、慢查询、缓存命中率等关键指标。
适用场景：数据库性能监控、故障排查、日常巡检。
当用户询问"MySQL状态"、"数据库连接数"、"慢查询"时使用。`
}

func (t *CheckMySQLStatusTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "mysql_host",
			Type:        "string",
			Description: "MySQL 服务器地址，默认 localhost",
			Required:    false,
		},
		{
			Name:        "mysql_port",
			Type:        "string",
			Description: "MySQL 端口，默认 3306",
			Required:    false,
		},
		{
			Name:        "mysql_user",
			Type:        "string",
			Description: "MySQL 用户名",
			Required:    true,
		},
		{
			Name:        "mysql_password",
			Type:        "string",
			Description: "MySQL 密码（可选）",
			Required:    false,
		},
	}
}

func (t *CheckMySQLStatusTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	mysqlHost := tool.GetStringParam(params, "mysql_host", "localhost")
	mysqlPort := tool.GetStringParam(params, "mysql_port", "3306")
	mysqlUser := tool.GetStringParam(params, "mysql_user", "")
	mysqlPassword := tool.GetStringParam(params, "mysql_password", "")

	if host == "" {
		return tool.NewErrorResult("请指定 host 参数"), nil
	}
	if mysqlUser == "" {
		return tool.NewErrorResult("请指定 mysql_user 参数"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	// 构建 MySQL 连接参数
	authPart := fmt.Sprintf("-u%s", mysqlUser)
	if mysqlPassword != "" {
		authPart = fmt.Sprintf("%s -p%s", authPart, mysqlPassword)
	}
	connPart := fmt.Sprintf("-h%s -P%s", mysqlHost, mysqlPort)

	// 检查 MySQL 状态
	statusCmd := fmt.Sprintf(`mysql %s %s -e "SHOW STATUS LIKE 'Threads_connected'; SHOW STATUS LIKE 'Threads_running'; SHOW STATUS LIKE 'Slow_queries'; SHOW STATUS LIKE 'Questions'; SHOW STATUS LIKE 'Uptime';"`, authPart, connPart)
	statusOut, err := ctx.SSH.Exec(host, statusCmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	// 检查进程列表
	processCmd := fmt.Sprintf(`mysql %s %s -e "SHOW PROCESSLIST;"`, authPart, connPart)
	processOut, err := ctx.SSH.Exec(host, processCmd)
	if err != nil {
		processOut = fmt.Sprintf("获取进程列表失败: %v", err)
	}

	return tool.NewResult(map[string]interface{}{
		"host":         host,
		"mysql_host":   mysqlHost,
		"mysql_port":   mysqlPort,
		"status":       strings.TrimSpace(statusOut),
		"process_list": strings.TrimSpace(processOut),
	}, fmt.Sprintf("成功获取 %s 上 MySQL 的状态信息", host)), nil
}

func NewCheckMySQLStatusTool() *CheckMySQLStatusTool { return &CheckMySQLStatusTool{} }

// CheckRedisStatusTool Redis 状态检查工具
type CheckRedisStatusTool struct{}

func (t *CheckRedisStatusTool) Name() string { return "check_redis_status" }

func (t *CheckRedisStatusTool) Description() string {
	return `检查 Redis 缓存状态。
返回 Redis 运行状态、内存使用、连接数、命中率等关键指标。
适用场景：缓存性能监控、内存使用分析、故障排查。
当用户询问"Redis状态"、"缓存使用"、"Redis内存"时使用。`
}

func (t *CheckRedisStatusTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "redis_host",
			Type:        "string",
			Description: "Redis 服务器地址，默认 localhost",
			Required:    false,
		},
		{
			Name:        "redis_port",
			Type:        "string",
			Description: "Redis 端口，默认 6379",
			Required:    false,
		},
		{
			Name:        "redis_password",
			Type:        "string",
			Description: "Redis 密码（可选）",
			Required:    false,
		},
	}
}

func (t *CheckRedisStatusTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	redisHost := tool.GetStringParam(params, "redis_host", "localhost")
	redisPort := tool.GetStringParam(params, "redis_port", "6379")
	redisPassword := tool.GetStringParam(params, "redis_password", "")

	if host == "" {
		return tool.NewErrorResult("请指定 host 参数"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	// 构建 Redis 连接参数
	authPart := ""
	if redisPassword != "" {
		authPart = fmt.Sprintf("-a %s", redisPassword)
	}
	connPart := fmt.Sprintf("-h %s -p %s", redisHost, redisPort)

	// 获取 Redis INFO
	infoCmd := fmt.Sprintf(`redis-cli %s %s INFO`, connPart, authPart)
	infoOut, err := ctx.SSH.Exec(host, infoCmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	// 获取客户端连接列表
	clientCmd := fmt.Sprintf(`redis-cli %s %s CLIENT LIST`, connPart, authPart)
	clientOut, err := ctx.SSH.Exec(host, clientCmd)
	if err != nil {
		clientOut = fmt.Sprintf("获取客户端列表失败: %v", err)
	}

	return tool.NewResult(map[string]interface{}{
		"host":        host,
		"redis_host":  redisHost,
		"redis_port":  redisPort,
		"info":        strings.TrimSpace(infoOut),
		"client_list": strings.TrimSpace(clientOut),
	}, fmt.Sprintf("成功获取 %s 上 Redis 的状态信息", host)), nil
}

func NewCheckRedisStatusTool() *CheckRedisStatusTool { return &CheckRedisStatusTool{} }

// BackupDatabaseTool 数据库备份工具
type BackupDatabaseTool struct{}

func (t *BackupDatabaseTool) Name() string { return "backup_database" }

func (t *BackupDatabaseTool) Description() string {
	return `备份 MySQL 或 PostgreSQL 数据库。
支持完整数据库备份，生成 SQL 转储文件。
适用场景：定期备份、迁移前备份、重要操作前备份。
当用户询问"备份数据库"、"导出数据"、"数据库迁移"时使用。`
}

func (t *BackupDatabaseTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "db_type",
			Type:        "string",
			Description: "数据库类型：mysql 或 postgresql",
			Required:    true,
		},
		{
			Name:        "db_name",
			Type:        "string",
			Description: "数据库名称",
			Required:    true,
		},
		{
			Name:        "db_user",
			Type:        "string",
			Description: "数据库用户名",
			Required:    true,
		},
		{
			Name:        "db_password",
			Type:        "string",
			Description: "数据库密码",
			Required:    true,
		},
		{
			Name:        "backup_path",
			Type:        "string",
			Description: "备份文件保存路径",
			Required:    true,
		},
	}
}

func (t *BackupDatabaseTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	dbType := tool.GetStringParam(params, "db_type", "")
	dbName := tool.GetStringParam(params, "db_name", "")
	dbUser := tool.GetStringParam(params, "db_user", "")
	dbPassword := tool.GetStringParam(params, "db_password", "")
	backupPath := tool.GetStringParam(params, "backup_path", "")

	if host == "" {
		return tool.NewErrorResult("请指定 host 参数"), nil
	}
	if dbType == "" {
		return tool.NewErrorResult("请指定 db_type 参数（mysql 或 postgresql）"), nil
	}
	if dbName == "" {
		return tool.NewErrorResult("请指定 db_name 参数"), nil
	}
	if dbUser == "" {
		return tool.NewErrorResult("请指定 db_user 参数"), nil
	}
	if dbPassword == "" {
		return tool.NewErrorResult("请指定 db_password 参数"), nil
	}
	if backupPath == "" {
		return tool.NewErrorResult("请指定 backup_path 参数"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	var backupCmd string
	switch strings.ToLower(dbType) {
	case "mysql":
		backupCmd = fmt.Sprintf(`mysqldump -u%s -p%s %s > %s`, dbUser, dbPassword, dbName, backupPath)
	case "postgresql":
		backupCmd = fmt.Sprintf(`PGPASSWORD=%s pg_dump -U %s -d %s -f %s`, dbPassword, dbUser, dbName, backupPath)
	default:
		return tool.NewErrorResult(fmt.Sprintf("不支持的数据库类型: %s（仅支持 mysql 或 postgresql）", dbType)), nil
	}

	output, err := ctx.SSH.Exec(host, backupCmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("备份失败: %v", err)), nil
	}

	// 检查备份文件是否创建成功
	checkCmd := fmt.Sprintf(`ls -lh %s`, backupPath)
	checkOut, err := ctx.SSH.Exec(host, checkCmd)
	if err != nil {
		return tool.NewResult(map[string]interface{}{
			"host":        host,
			"db_type":     dbType,
			"db_name":     dbName,
			"backup_path": backupPath,
			"output":      strings.TrimSpace(output),
			"warning":     "备份命令执行完成，但无法验证文件",
		}, fmt.Sprintf("数据库 %s 备份完成", dbName)), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":        host,
		"db_type":     dbType,
		"db_name":     dbName,
		"backup_path": backupPath,
		"file_info":   strings.TrimSpace(checkOut),
		"output":      strings.TrimSpace(output),
	}, fmt.Sprintf("成功备份数据库 %s 到 %s", dbName, backupPath)), nil
}

func NewBackupDatabaseTool() *BackupDatabaseTool { return &BackupDatabaseTool{} }
