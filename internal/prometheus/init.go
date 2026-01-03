package prometheus

import (
	"context"
	"fmt"
	"time"

	"ai-ops/internal/config"
	"ai-ops/internal/tool"
	"go.uber.org/zap"
)

// InitPrometheusIntegration 初始化 Prometheus 集成
func InitPrometheusIntegration(cfg *config.Config, toolRegistry *tool.Registry, logger *zap.Logger) error {
	if !cfg.Prometheus.Enabled {
		logger.Info("Prometheus 集成未启用")
		return nil
	}

	if cfg.Prometheus.Address == "" {
		logger.Warn("Prometheus 地址未配置")
		return nil
	}

	// 创建 Prometheus 客户端
	promClient, err := NewPrometheusClient(cfg.Prometheus.Address)
	if err != nil {
		logger.Error("创建 Prometheus 客户端失败", zap.Error(err))
		return fmt.Errorf("创建 Prometheus 客户端失败: %w", err)
	}

	logger.Info("Prometheus 客户端初始化完成",
		zap.String("address", cfg.Prometheus.Address),
		zap.Int("timeout", cfg.Prometheus.Timeout),
	)

	// 验证 Prometheus 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata, err := promClient.GetMetricMetadata(ctx)
	if err != nil {
		logger.Warn("无法连接到 Prometheus 服务器", zap.Error(err))
	} else {
		logger.Info("Prometheus 连接验证成功",
			zap.Int("metric_count", len(metadata)),
		)
	}

	return nil
}

// 在 cmd/server/main.go 中的使用示例：
/*
func main() {
	// ... 其他初始化代码 ...

	// 初始化 Prometheus 集成
	if err := InitPrometheusIntegration(cfg, toolRegistry, logger); err != nil {
		logger.Warn("Prometheus 集成初始化失败", zap.Error(err))
		// 继续启动，Prometheus 是可选的
	}

	// 传递 Prometheus 地址给路由
	router := api.NewRouter(api.RouterConfig{
		// ... 其他配置 ...
		PrometheusAddr: cfg.Prometheus.Address,
	})

	// ... 启动服务 ...
}
*/
