package handler

import (
	"ai-ops/internal/feedback"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FeedbackHandler 反馈处理器
type FeedbackHandler struct {
	feedbackService *feedback.Service
}

// NewFeedbackHandler 创建反馈处理器
func NewFeedbackHandler(feedbackService *feedback.Service) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackService: feedbackService,
	}
}

// SubmitFeedback 提交反馈
// POST /api/feedback
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		SessionID string `json:"session_id"`
		Question  string `json:"question"`
		Answer    string `json:"answer"`
		Rating    int    `json:"rating" binding:"required,oneof=-1 1"` // 1=好, -1=差
		Comment   string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	fb := &feedback.Feedback{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		MessageID: req.MessageID,
		Question:  req.Question,
		Answer:    req.Answer,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := h.feedbackService.RecordFeedback(c.Request.Context(), fb); err != nil {
		InternalError(c, "保存反馈失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":      fb.ID,
		"message": "反馈已记录",
	})
}

// GetFeedback 获取反馈
// GET /api/feedback/:message_id
func (h *FeedbackHandler) GetFeedback(c *gin.Context) {
	messageID := c.Param("message_id")
	if messageID == "" {
		ParamError(c, "message_id 不能为空")
		return
	}

	fb, err := h.feedbackService.GetByMessageID(c.Request.Context(), messageID)
	if err != nil {
		NotFound(c, "反馈不存在")
		return
	}

	Success(c, fb)
}
