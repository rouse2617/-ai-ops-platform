package operation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// ConfirmationManager 确认管理器
type ConfirmationManager struct {
	repo     *repository.OperationRepository
	pending  map[string]*PendingConfirmation
	mu       sync.RWMutex
	notifier Notifier
}

// PendingConfirmation 待确认的操作
type PendingConfirmation struct {
	Confirmation *model.OperationConfirmation
	ResultChan   chan bool
	Timer        *time.Timer
}

// Notifier 通知接口
type Notifier interface {
	NotifyConfirmationRequest(confirmation *model.OperationConfirmation) error
}

// NewConfirmationManager 创建确认管理器
func NewConfirmationManager(repo *repository.OperationRepository, notifier Notifier) *ConfirmationManager {
	return &ConfirmationManager{
		repo:     repo,
		pending:  make(map[string]*PendingConfirmation),
		notifier: notifier,
	}
}

// RequestConfirmation 请求确认
func (m *ConfirmationManager) RequestConfirmation(ctx context.Context, confirmation *model.OperationConfirmation) (bool, error) {
	// 保存到数据库
	if err := m.repo.Create(confirmation); err != nil {
		return false, fmt.Errorf("保存确认请求失败: %w", err)
	}

	// 创建待确认记录
	pending := &PendingConfirmation{
		Confirmation: confirmation,
		ResultChan:   make(chan bool, 1),
	}

	// 设置超时定时器
	timeout := time.Until(confirmation.TimeoutAt)
	pending.Timer = time.AfterFunc(timeout, func() {
		m.handleTimeout(confirmation.ID)
	})

	// 添加到待确认列表
	m.mu.Lock()
	m.pending[confirmation.ID] = pending
	m.mu.Unlock()

	// 发送通知
	if m.notifier != nil {
		if err := m.notifier.NotifyConfirmationRequest(confirmation); err != nil {
			logger.Error("发送确认通知失败", zap.Error(err))
		}
	}

	// 等待结果
	select {
	case approved := <-pending.ResultChan:
		return approved, nil
	case <-ctx.Done():
		m.cleanup(confirmation.ID)
		return false, ctx.Err()
	}
}

// Approve 批准操作
func (m *ConfirmationManager) Approve(confirmationID string, confirmedBy string) error {
	m.mu.Lock()
	pending, exists := m.pending[confirmationID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("确认请求不存在或已过期")
	}

	// 停止超时定时器
	pending.Timer.Stop()

	// 更新数据库
	now := time.Now()
	pending.Confirmation.Status = model.ConfirmationStatusApproved
	pending.Confirmation.ConfirmedAt = &now
	pending.Confirmation.ConfirmedBy = confirmedBy

	if err := m.repo.Update(pending.Confirmation); err != nil {
		return fmt.Errorf("更新确认状态失败: %w", err)
	}

	// 发送结果
	select {
	case pending.ResultChan <- true:
	default:
	}

	// 清理
	m.cleanup(confirmationID)

	logger.Info("操作已批准",
		zap.String("confirmation_id", confirmationID),
		zap.String("confirmed_by", confirmedBy),
	)

	return nil
}

// Reject 拒绝操作
func (m *ConfirmationManager) Reject(confirmationID string, confirmedBy string, reason string) error {
	m.mu.Lock()
	pending, exists := m.pending[confirmationID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("确认请求不存在或已过期")
	}

	// 停止超时定时器
	pending.Timer.Stop()

	// 更新数据库
	now := time.Now()
	pending.Confirmation.Status = model.ConfirmationStatusRejected
	pending.Confirmation.ConfirmedAt = &now
	pending.Confirmation.ConfirmedBy = confirmedBy
	if reason != "" {
		pending.Confirmation.Reason = reason
	}

	if err := m.repo.Update(pending.Confirmation); err != nil {
		return fmt.Errorf("更新确认状态失败: %w", err)
	}

	// 发送结果
	select {
	case pending.ResultChan <- false:
	default:
	}

	// 清理
	m.cleanup(confirmationID)

	logger.Info("操作已拒绝",
		zap.String("confirmation_id", confirmationID),
		zap.String("confirmed_by", confirmedBy),
		zap.String("reason", reason),
	)

	return nil
}

// handleTimeout 处理超时
func (m *ConfirmationManager) handleTimeout(confirmationID string) {
	m.mu.Lock()
	pending, exists := m.pending[confirmationID]
	m.mu.Unlock()

	if !exists {
		return
	}

	// 更新数据库
	pending.Confirmation.Status = model.ConfirmationStatusTimeout

	if err := m.repo.Update(pending.Confirmation); err != nil {
		logger.Error("更新超时状态失败", zap.Error(err))
	}

	// 发送结果
	select {
	case pending.ResultChan <- false:
	default:
	}

	// 清理
	m.cleanup(confirmationID)

	logger.Warn("确认请求超时",
		zap.String("confirmation_id", confirmationID),
	)
}

// cleanup 清理待确认记录
func (m *ConfirmationManager) cleanup(confirmationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pending, exists := m.pending[confirmationID]; exists {
		close(pending.ResultChan)
		delete(m.pending, confirmationID)
	}
}

// GetPending 获取待确认列表
func (m *ConfirmationManager) GetPending(sessionID string) ([]*model.OperationConfirmation, error) {
	return m.repo.FindBySessionAndStatus(sessionID, model.ConfirmationStatusPending)
}

// GetByID 根据 ID 获取确认记录
func (m *ConfirmationManager) GetByID(confirmationID string) (*model.OperationConfirmation, error) {
	return m.repo.FindByID(confirmationID)
}

// List 列出确认记录
func (m *ConfirmationManager) List(sessionID string, limit int) ([]*model.OperationConfirmation, error) {
	return m.repo.FindBySession(sessionID, limit)
}
