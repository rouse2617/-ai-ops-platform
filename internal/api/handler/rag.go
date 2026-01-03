package handler

import (
	"net/http"

	"ai-ops/internal/rag"
	"ai-ops/internal/service"

	"github.com/gin-gonic/gin"
)

// RAGHandler RAG 知识库处理器
type RAGHandler struct {
	ragService  *rag.Service
	chatService *service.ChatService
}

// NewRAGHandler 创建 RAG 处理器
func NewRAGHandler(ragService *rag.Service, chatService *service.ChatService) *RAGHandler {
	return &RAGHandler{
		ragService:  ragService,
		chatService: chatService,
	}
}

// UploadDocument 上传文档
// POST /api/v1/knowledge/documents
func (h *RAGHandler) UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	// 解析元数据
	_ = c.PostForm("category")
	_ = c.PostForm("source")
	_ = c.PostFormArray("tags")

	// TODO: 解析文件并索引
	c.JSON(http.StatusOK, gin.H{
		"message": "文档上传成功",
		"file":    file.Filename,
	})
}

// SearchKnowledge 检索知识
// POST /api/v1/knowledge/search
func (h *RAGHandler) SearchKnowledge(c *gin.Context) {
	var req struct {
		Query   string            `json:"query" binding:"required"`
		TopK    int               `json:"top_k"`
		Filters map[string]string `json:"filters"`
		Rerank  bool              `json:"rerank"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	// 执行检索
	knowledge, err := h.ragService.Search(c.Request.Context(), rag.SearchQuery{
		Query:   req.Query,
		TopK:    req.TopK,
		Filters: req.Filters,
		Rerank:  req.Rerank,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, knowledge)
}

// GetAlertKnowledge 获取告警相关知识
// GET /api/v1/alerts/:id/knowledge
func (h *RAGHandler) GetAlertKnowledge(c *gin.Context) {
	alertID := c.Param("id")

	// TODO: 从告警系统获取告警信息
	// alert, err := h.alertService.GetAlert(alertID)

	// 模拟告警数据
	alert := map[string]interface{}{
		"id":      alertID,
		"service": "api-service",
		"message": "CPU usage above 90%",
	}

	// RAG 检索相关知识
	knowledge, err := h.ragService.SearchByAlert(c.Request.Context(), alert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alert":      alert,
		"knowledge":  knowledge,
		"confidence": knowledge.Confidence,
	})
}

// ListDocuments 列出文档
// GET /api/v1/knowledge/documents
func (h *RAGHandler) ListDocuments(c *gin.Context) {
	// TODO: 实现文档列表查询
	c.JSON(http.StatusOK, gin.H{
		"documents": []interface{}{},
		"total":     0,
	})
}

// DeleteDocument 删除文档
// DELETE /api/v1/knowledge/documents/:id
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	docID := c.Param("id")

	if err := h.ragService.DeleteDocument(c.Request.Context(), docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档删除成功"})
}
