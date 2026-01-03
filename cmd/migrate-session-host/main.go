package main

import (
	"ai-ops/internal/model"
	"ai-ops/pkg/logger"
	"flag"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := flag.String("db", "./data/aiops.db", "数据库路径")
	dryRun := flag.Bool("dry-run", false, "模拟运行，不实际修改数据")
	flag.Parse()

	logger.Init(logger.Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})
	defer logger.Sync()

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 自动迁移 Message 表（添加 host_id 字段）
	if err := db.AutoMigrate(&model.Message{}); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	logger.Info("Message 表迁移完成，已添加 host_id 字段")

	// 创建索引
	if !*dryRun {
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_messages_session_host
			ON messages(session_id, host_id)
		`).Error; err != nil {
			logger.Error("创建索引失败", zap.Error(err))
		} else {
			logger.Info("索引创建成功: idx_messages_session_host")
		}
	}

	// 统计信息
	var totalMessages int64
	db.Model(&model.Message{}).Count(&totalMessages)

	var messagesWithHost int64
	db.Model(&model.Message{}).Where("host_id != ''").Count(&messagesWithHost)

	fmt.Println("\n========== 迁移统计 ==========")
	fmt.Printf("总消息数:           %d\n", totalMessages)
	fmt.Printf("已关联主机消息数:   %d\n", messagesWithHost)
	fmt.Printf("未关联主机消息数:   %d\n", totalMessages-messagesWithHost)

	if *dryRun {
		fmt.Println("\n【注意】这是模拟运行，未实际修改数据")
	}

	fmt.Println("==============================")
}
