package operation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"ai-ops/internal/model"
	"ai-ops/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DangerPattern 高危命令模式
type DangerPattern struct {
	Pattern   *regexp.Regexp
	RiskLevel string
	Reason    string
}

// DangerInterceptor 高危操作拦截器
type DangerInterceptor struct {
	patterns     []DangerPattern
	confirmMgr   *ConfirmationManager
	mu           sync.RWMutex
}

// NewDangerInterceptor 创建拦截器
func NewDangerInterceptor(confirmMgr *ConfirmationManager) *DangerInterceptor {
	interceptor := &DangerInterceptor{
		confirmMgr: confirmMgr,
		patterns:   make([]DangerPattern, 0),
	}
	interceptor.loadDefaultPatterns()
	return interceptor
}

// loadDefaultPatterns 加载默认高危命令模式
func (d *DangerInterceptor) loadDefaultPatterns() {
	patterns := []struct {
		pattern   string
		riskLevel string
		reason    string
	}{
		{`\brm\s+(-rf?|--recursive|--force)\s+/`, model.RiskLevelCritical, "删除根目录或系统目录"},
		{`\brm\s+.*\s+/`, model.RiskLevelCritical, "删除根目录"},
		{`\bdd\s+.*of=/dev/(sd|hd|nvme)`, model.RiskLevelCritical, "直接写入磁盘设备"},
		{`\bmkfs\b`, model.RiskLevelCritical, "格式化文件系统"},
		{`\bshutdown\b`, model.RiskLevelHigh, "关机操作"},
		{`\breboot\b`, model.RiskLevelHigh, "重启操作"},
		{`\bkillall\b`, model.RiskLevelHigh, "批量终止进程"},
		{`\biptables\s+(-F|--flush)`, model.RiskLevelHigh, "清空防火墙规则"},
		{`\bchmod\s+777`, model.RiskLevelMedium, "设置过于宽松的权限"},
		{`\bchown\s+.*\s+/`, model.RiskLevelMedium, "修改根目录权限"},
		{`>\s*/dev/(sd|hd|nvme)`, model.RiskLevelCritical, "重定向到磁盘设备"},
		{`\bdropdb\b`, model.RiskLevelHigh, "删除数据库"},
		{`\btruncate\b.*\s+/`, model.RiskLevelHigh, "清空系统文件"},
	}

	for _, p := range patterns {
		re, err := regexp.Compile(p.pattern)
		if err != nil {
			logger.Error("编译高危命令模式失败", zap.String("pattern", p.pattern), zap.Error(err))
			continue
		}
		d.patterns = append(d.patterns, DangerPattern{
			Pattern:   re,
			RiskLevel: p.riskLevel,
			Reason:    p.reason,
		})
	}
}

// CheckCommand 检查命令是否高危
func (d *DangerInterceptor) CheckCommand(cmd string) (isDangerous bool, riskLevel string, reason string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	cmd = strings.TrimSpace(cmd)
	for _, pattern := range d.patterns {
		if pattern.Pattern.MatchString(cmd) {
			return true, pattern.RiskLevel, pattern.Reason
		}
	}
	return false, "", ""
}

// InterceptAndConfirm 拦截并请求确认
func (d *DangerInterceptor) InterceptAndConfirm(ctx context.Context, sessionID, hostID, command string) error {
	isDangerous, riskLevel, reason := d.CheckCommand(command)
	if !isDangerous {
		return nil
	}

	logger.Warn("检测到高危命令",
		zap.String("session_id", sessionID),
		zap.String("host_id", hostID),
		zap.String("command", command),
		zap.String("risk_level", riskLevel),
		zap.String("reason", reason),
	)

	// 创建确认请求
	confirmation := &model.OperationConfirmation{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		HostID:      hostID,
		Command:     command,
		RiskLevel:   riskLevel,
		Status:      model.ConfirmationStatusPending,
		RequestedAt: time.Now(),
		TimeoutAt:   time.Now().Add(30 * time.Second),
		Reason:      reason,
		CreatedAt:   time.Now(),
	}

	// 等待用户确认
	approved, err := d.confirmMgr.RequestConfirmation(ctx, confirmation)
	if err != nil {
		return fmt.Errorf("确认请求失败: %w", err)
	}

	if !approved {
		return fmt.Errorf("操作被拒绝: %s", reason)
	}

	return nil
}

// AddPattern 添加自定义高危模式
func (d *DangerInterceptor) AddPattern(pattern string, riskLevel string, reason string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.patterns = append(d.patterns, DangerPattern{
		Pattern:   re,
		RiskLevel: riskLevel,
		Reason:    reason,
	})

	return nil
}

// RemovePattern 移除模式
func (d *DangerInterceptor) RemovePattern(pattern string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, p := range d.patterns {
		if p.Pattern.String() == pattern {
			d.patterns = append(d.patterns[:i], d.patterns[i+1:]...)
			break
		}
	}
}

// ListPatterns 列出所有模式
func (d *DangerInterceptor) ListPatterns() []DangerPattern {
	d.mu.RLock()
	defer d.mu.RUnlock()

	patterns := make([]DangerPattern, len(d.patterns))
	copy(patterns, d.patterns)
	return patterns
}
