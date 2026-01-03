package api

import (
	"ai-ops/internal/api/handler"
	"ai-ops/internal/api/middleware"
	"ai-ops/internal/auth"
	"ai-ops/internal/cache"
	"ai-ops/internal/config"
	"ai-ops/internal/llm"
	"ai-ops/internal/repository"
	"ai-ops/internal/security"
	"ai-ops/internal/service"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"

	"github.com/gin-gonic/gin"
)

// RouterConfig 路由配置
type RouterConfig struct {
	SSHPool      *ssh.Pool
	ToolRegistry *tool.Registry
	PolicyStore  *security.PolicyStore
	AuditLogger  *security.AuditLogger
	HostRepo     repository.HostRepository
	SessionRepo  repository.SessionRepository
	GroupRepo    repository.GroupRepository
	ConfigRepo   repository.ConfigRepository
	AnalysisRepo repository.AnalysisRepository
	LLMClient    *llm.OpenAIClient
	Cache        cache.Cache // 可选的缓存
	Version      string
	Mode         string // debug / release
	Config       *config.Config
}

// NewRouter 创建路由
func NewRouter(cfg RouterConfig) *gin.Engine {
	// 设置运行模式
	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 中间件
	r.Use(RecoveryMiddleware())
	r.Use(LoggerMiddleware())
	r.Use(CORSMiddleware())

	// 创建 JWT 管理器（如果启用了认证）
	var jwtManager *auth.JWTManager
	var authMiddleware gin.HandlerFunc

	if cfg.Config != nil && cfg.Config.Auth.Enabled {
		jwtManager = auth.NewJWTManager(
			cfg.Config.Auth.Secret,
			cfg.Config.Auth.TokenDuration,
			cfg.Config.Auth.Issuer,
		)
		authMiddleware = middleware.AuthMiddleware(jwtManager)
	} else {
		// 如果未启用认证，使用空中间件
		authMiddleware = func(c *gin.Context) { c.Next() }
	}

	// 创建 LLM 客户端
	llmClient, err := llm.NewClient(llm.ProviderConfig{
		Type:      llm.ProviderType(cfg.Config.LLM.Provider),
		Endpoint:  cfg.Config.LLM.Endpoint,
		Model:     cfg.Config.LLM.Model,
		APIKey:    cfg.Config.LLM.APIKey,
		Timeout:   cfg.Config.LLM.Timeout,
		MaxTokens: cfg.Config.LLM.MaxTokens,
	})
	if err != nil {
		panic("创建 LLM 客户端失败: " + err.Error())
	}

	// 创建 Service 层
	chatService := service.NewChatService(llmClient, cfg.ToolRegistry, cfg.SSHPool, cfg.SessionRepo)
	hostService := service.NewHostService(cfg.SSHPool, cfg.HostRepo, cfg.GroupRepo, cfg.Cache)

	// 创建 handlers
	chatHandler := handler.NewChatHandler(chatService)
	hostHandler := handler.NewHostHandler(hostService)
	systemHandler := handler.NewSystemHandler(cfg.Version, cfg.PolicyStore, cfg.AuditLogger, cfg.ConfigRepo)
	analysisHandler := handler.NewAnalysisHandler(cfg.LLMClient, cfg.AnalysisRepo)
	internalSSHHandler := handler.NewInternalSSHHandler(cfg.SSHPool, cfg.HostRepo)

	// API 路由组
	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		public := api.Group("")
		{
			// 健康检查
			public.GET("/health", systemHandler.Health)

			// 认证接口
			if jwtManager != nil {
				authHandler := handler.NewAuthHandler(jwtManager)
				public.POST("/auth/login", authHandler.Login)
				public.POST("/auth/refresh", authHandler.RefreshToken)
				public.POST("/auth/logout", authHandler.Logout)
				public.GET("/auth/validate", authHandler.ValidateToken)
			}
		}

		// 需要认证的接口
		protected := api.Group("")
		if jwtManager != nil {
			protected.Use(authMiddleware)
		}

		{
			// 对话 API
			chat := protected.Group("/chat")
			{
				chat.POST("", chatHandler.Chat)
				chat.POST("/send", chatHandler.Chat) // 兼容前端 /chat/send 调用
				chat.POST("/stream", chatHandler.StreamChat)
				chat.GET("/sessions", chatHandler.GetSessions)
				chat.POST("/sessions", chatHandler.CreateSession)
				chat.GET("/history/:session_id", chatHandler.GetHistory)
				chat.DELETE("/sessions/:id", chatHandler.DeleteSession)
			}

			// 主机管理 API
			hosts := protected.Group("/hosts")
			{
				hosts.GET("", hostHandler.ListHosts)
				hosts.GET("/all", hostHandler.GetAllHosts)
				hosts.POST("", hostHandler.CreateHost)
				hosts.POST("/batch-delete", hostHandler.BatchDeleteHosts)
				hosts.POST("/import", hostHandler.ImportHosts)
				hosts.GET("/:id", hostHandler.GetHost)
				hosts.PUT("/:id", hostHandler.UpdateHost)
				hosts.DELETE("/:id", hostHandler.DeleteHost)
				hosts.POST("/:id/test", hostHandler.TestConnection)
			}

			// 分组 API
			groups := protected.Group("/groups")
			{
				groups.GET("", hostHandler.GetGroups)
				groups.POST("", hostHandler.CreateGroup)
				groups.DELETE("/:name", hostHandler.DeleteGroup)
			}

			// 系统 API
			system := protected.Group("/system")
			{
				system.GET("/info", systemHandler.GetSystemInfo)
				system.GET("/config", systemHandler.GetConfig)
				system.PUT("/config", systemHandler.UpdateConfig)
				system.GET("/command-policy", systemHandler.GetCommandPolicy)
				system.PUT("/command-policy", systemHandler.UpdateCommandPolicy)
				system.GET("/audit/commands", systemHandler.GetCommandAudit)
			}

			// AI分析 API
			analysis := protected.Group("/analysis")
			{
				analysis.POST("/analyze", analysisHandler.Analyze)
				analysis.GET("/history", analysisHandler.GetHistory)
				analysis.GET("/:id", analysisHandler.GetAnalysis)
				analysis.DELETE("/:id", analysisHandler.DeleteAnalysis)
			}
		}

		// 内部服务接口（仅供 Node.js Agent Service 调用）
		internal := api.Group("/internal")
		{
			internal.GET("/hosts/:ip/credentials", internalSSHHandler.GetHostCredentials)
		}
	}

	// 静态文件服务（前端）
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/", "./web/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	return r
}
