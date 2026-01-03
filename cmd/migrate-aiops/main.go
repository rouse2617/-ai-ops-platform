package main

import (
	"fmt"
	"log"

	"ai-ops/internal/config"
	"ai-ops/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("开始数据库迁移...")

	// 自动迁移现有模型
	if err := db.AutoMigrate(
		&model.Host{},
		&model.Session{},
		&model.Message{},
		&model.Analysis{},
	); err != nil {
		log.Fatalf("迁移现有模型失败: %v", err)
	}
	fmt.Println("✓ 现有模型迁移完成")

	// 迁移新增模型
	if err := db.AutoMigrate(
		&model.OperationConfirmation{},
		&model.HealthCheck{},
		&model.TrendPrediction{},
		&model.PrometheusAlert{},
	); err != nil {
		log.Fatalf("迁移新增模型失败: %v", err)
	}
	fmt.Println("✓ 新增模型迁移完成")

	// 更新 sessions 表结构
	if err := migrateSessionsTable(db); err != nil {
		log.Printf("警告: 更新 sessions 表失败: %v", err)
	} else {
		fmt.Println("✓ Sessions 表更新完成")
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		log.Printf("警告: 创建索引失败: %v", err)
	} else {
		fmt.Println("✓ 索引创建完成")
	}

	fmt.Println("\n数据库迁移完成!")
	fmt.Println("\n新增表:")
	fmt.Println("  - operation_confirmations (操作确认)")
	fmt.Println("  - health_checks (健康检查)")
	fmt.Println("  - trend_predictions (趋势预测)")
	fmt.Println("  - prometheus_alerts (Prometheus 告警)")
}

// migrateSessionsTable 更新 sessions 表
func migrateSessionsTable(db *gorm.DB) error {
	// 检查列是否存在
	type columnInfo struct {
		Name string
	}

	var columns []columnInfo
	if err := db.Raw("PRAGMA table_info(sessions)").Scan(&columns).Error; err != nil {
		return err
	}

	hasContext := false
	hasLastActivity := false
	hasHostIDs := false

	for _, col := range columns {
		switch col.Name {
		case "context":
			hasContext = true
		case "last_activity":
			hasLastActivity = true
		case "host_ids":
			hasHostIDs = true
		}
	}

	// 添加缺失的列
	if !hasContext {
		if err := db.Exec("ALTER TABLE sessions ADD COLUMN context TEXT").Error; err != nil {
			return fmt.Errorf("添加 context 列失败: %w", err)
		}
	}

	if !hasLastActivity {
		if err := db.Exec("ALTER TABLE sessions ADD COLUMN last_activity DATETIME").Error; err != nil {
			return fmt.Errorf("添加 last_activity 列失败: %w", err)
		}
	}

	if !hasHostIDs {
		if err := db.Exec("ALTER TABLE sessions ADD COLUMN host_ids TEXT").Error; err != nil {
			return fmt.Errorf("添加 host_ids 列失败: %w", err)
		}
	}

	return nil
}

// createIndexes 创建索引
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_operation_confirmations_session ON operation_confirmations(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_operation_confirmations_status ON operation_confirmations(status)",
		"CREATE INDEX IF NOT EXISTS idx_operation_confirmations_timeout ON operation_confirmations(timeout_at)",
		"CREATE INDEX IF NOT EXISTS idx_health_checks_host ON health_checks(host_id)",
		"CREATE INDEX IF NOT EXISTS idx_health_checks_checked_at ON health_checks(checked_at)",
		"CREATE INDEX IF NOT EXISTS idx_trend_predictions_host ON trend_predictions(host_id)",
		"CREATE INDEX IF NOT EXISTS idx_trend_predictions_time ON trend_predictions(prediction_time)",
		"CREATE INDEX IF NOT EXISTS idx_prometheus_alerts_host ON prometheus_alerts(host_id)",
		"CREATE INDEX IF NOT EXISTS idx_prometheus_alerts_status ON prometheus_alerts(status)",
		"CREATE INDEX IF NOT EXISTS idx_prometheus_alerts_starts_at ON prometheus_alerts(starts_at)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	return nil
}
