package monitor

import (
	"context"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/tool/builtin"
	"go.uber.org/zap"
)

// AlertManager manages alerts and triggers code correlation analysis
type AlertManager struct {
	notifier   *AlertNotifier
	correlator *CodeChangeCorrelator
	logger     *zap.Logger
}

// NewAlertManager creates a new alert manager
func NewAlertManager(notifier *AlertNotifier, gitTracker *builtin.GitChangeTracker, logger *zap.Logger) *AlertManager {
	return &AlertManager{
		notifier:   notifier,
		correlator: NewCodeChangeCorrelator(gitTracker),
		logger:     logger,
	}
}

// RegisterServiceMapping registers service to code mapping
func (am *AlertManager) RegisterServiceMapping(mapping *model.ServiceCodeMapping) {
	am.correlator.RegisterServiceMapping(mapping)
}

// HandleAlert processes an alert and triggers correlation analysis
func (am *AlertManager) HandleAlert(ctx context.Context, alert *Alert) error {
	// Send basic alert notification
	if err := am.notifier.Notify(alert); err != nil {
		am.logger.Error("Failed to notify alert", zap.Error(err))
	}

	// Trigger code correlation analysis in background
	go am.analyzeCodeCorrelation(context.Background(), alert)

	return nil
}

// analyzeCodeCorrelation performs code correlation analysis
func (am *AlertManager) analyzeCodeCorrelation(ctx context.Context, alert *Alert) {
	correlation, err := am.correlator.AnalyzeAlert(ctx, alert)
	if err != nil {
		am.logger.Error("Code correlation analysis failed",
			zap.String("alert_id", alert.ID),
			zap.Error(err))
		return
	}

	if correlation == nil || correlation.CorrelationScore < 0.3 {
		am.logger.Debug("No significant code correlation found",
			zap.String("alert_id", alert.ID))
		return
	}

	// Send correlation notification
	notification := am.buildCorrelationNotification(correlation)
	if err := am.notifier.NotifyCodeCorrelation(notification); err != nil {
		am.logger.Error("Failed to notify code correlation",
			zap.String("alert_id", alert.ID),
			zap.Error(err))
	}

	am.logger.Info("Code correlation detected",
		zap.String("alert_id", alert.ID),
		zap.Float64("score", correlation.CorrelationScore),
		zap.Int("commits", len(correlation.Commits)))
}

// buildCorrelationNotification builds a user-friendly notification
func (am *AlertManager) buildCorrelationNotification(correlation *model.CodeChangeCorrelation) map[string]interface{} {
	message := am.generateCorrelationMessage(correlation)

	return map[string]interface{}{
		"alert_id":          correlation.AlertID,
		"alert_name":        correlation.AlertName,
		"service":           correlation.Service,
		"correlation_score": correlation.CorrelationScore,
		"message":           message,
		"commits":           correlation.Commits,
		"suspicious_files":  correlation.SuspiciousFiles,
		"time_window":       correlation.TimeWindow,
		"timestamp":         time.Now(),
	}
}

// generateCorrelationMessage generates a human-readable message
func (am *AlertManager) generateCorrelationMessage(correlation *model.CodeChangeCorrelation) string {
	if len(correlation.SuspiciousFiles) == 0 {
		return "指标异常！检测到代码变更，但未找到明确的可疑文件。"
	}

	topFile := correlation.SuspiciousFiles[0]
	commitCount := len(correlation.Commits)

	var timeDesc string
	timeDiff := correlation.TimeWindow.AlertTime.Sub(correlation.Commits[0].Timestamp)
	if timeDiff < 5*time.Minute {
		timeDesc = "刚刚"
	} else if timeDiff < 15*time.Minute {
		timeDesc = "10分钟前"
	} else {
		timeDesc = "30分钟内"
	}

	if commitCount == 1 {
		return "指标异常！我发现" + timeDesc + "有一条新的代码合并，改动了 " + topFile.FilePath +
			"。这次指标波动极大概率是由代码变更引起的，是否查看改动对比？"
	}

	return "指标异常！我��现" + timeDesc + "有 " + string(rune(commitCount+'0')) + " 条代码合并，其中 " +
		topFile.FilePath + " 的改动最可疑。这次指标波动极大概率是由代码变更引起的，是否查看改动对比？"
}
