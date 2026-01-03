package service

import (
	"context"
	"fmt"

	"ai-ops/internal/operation"
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"
)

// OperationService 操作服务
type OperationService struct {
	interceptor *operation.DangerInterceptor
	confirmMgr  *operation.ConfirmationManager
	sshPool     *ssh.Pool
	hostRepo    repository.HostRepository
}

// NewOperationService 创建操作服务
func NewOperationService(
	interceptor *operation.DangerInterceptor,
	confirmMgr *operation.ConfirmationManager,
	sshPool *ssh.Pool,
	hostRepo repository.HostRepository,
) *OperationService {
	return &OperationService{
		interceptor: interceptor,
		confirmMgr:  confirmMgr,
		sshPool:     sshPool,
		hostRepo:    hostRepo,
	}
}

// ExecuteCommand 执行命令（带拦截）
func (s *OperationService) ExecuteCommand(ctx context.Context, sessionID, hostID, command string) (string, error) {
	// 检查并拦截高危命令
	if err := s.interceptor.InterceptAndConfirm(ctx, sessionID, hostID, command); err != nil {
		return "", fmt.Errorf("命令被拦截: %w", err)
	}

	// 获取主机信息
	host, err := s.hostRepo.GetByID(hostID)
	if err != nil {
		return "", fmt.Errorf("主机不存在: %w", err)
	}

	// 执行命令
	output, err := s.sshPool.Exec(host.Name, command)
	if err != nil {
		return "", fmt.Errorf("执行失败: %w", err)
	}

	return output, nil
}

// BatchExecuteCommand 批量执行命令
func (s *OperationService) BatchExecuteCommand(ctx context.Context, sessionID string, hostIDs []string, command string) (map[string]string, error) {
	results := make(map[string]string)

	for _, hostID := range hostIDs {
		output, err := s.ExecuteCommand(ctx, sessionID, hostID, command)
		if err != nil {
			results[hostID] = fmt.Sprintf("错误: %v", err)
		} else {
			results[hostID] = output
		}
	}

	return results, nil
}

// CheckCommand 检查命令是否高危
func (s *OperationService) CheckCommand(command string) (isDangerous bool, riskLevel string, reason string) {
	return s.interceptor.CheckCommand(command)
}
