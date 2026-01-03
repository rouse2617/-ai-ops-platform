package handler

import (
	"ai-ops/internal/operation"
	"ai-ops/internal/service"

	"github.com/gin-gonic/gin"
)

// OperationHandler 操作处理器
type OperationHandler struct {
	operationService *service.OperationService
	confirmMgr       *operation.ConfirmationManager
}

// NewOperationHandler 创建操作处理器
func NewOperationHandler(operationService *service.OperationService, confirmMgr *operation.ConfirmationManager) *OperationHandler {
	return &OperationHandler{
		operationService: operationService,
		confirmMgr:       confirmMgr,
	}
}

// GetConfirmation 获取确认请求详情
// GET /api/operations/confirmations/:id
func (h *OperationHandler) GetConfirmation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	confirmation, err := h.confirmMgr.GetByID(id)
	if err != nil {
		NotFound(c, "确认请求不存在")
		return
	}

	Success(c, confirmation)
}

// ApproveConfirmation 批准操作
// POST /api/operations/confirmations/:id/approve
func (h *OperationHandler) ApproveConfirmation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	var req struct {
		ConfirmedBy string `json:"confirmed_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body
	}

	if req.ConfirmedBy == "" {
		req.ConfirmedBy = "user"
	}

	if err := h.confirmMgr.Approve(id, req.ConfirmedBy); err != nil {
		InternalError(c, "批准操作失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "操作已批准", nil)
}

// RejectConfirmation 拒绝操作
// POST /api/operations/confirmations/:id/reject
func (h *OperationHandler) RejectConfirmation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	var req struct {
		ConfirmedBy string `json:"confirmed_by"`
		Reason      string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body
	}

	if req.ConfirmedBy == "" {
		req.ConfirmedBy = "user"
	}

	if err := h.confirmMgr.Reject(id, req.ConfirmedBy, req.Reason); err != nil {
		InternalError(c, "拒绝操作失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "操作已拒绝", nil)
}

// ListConfirmations 列出确认请求
// GET /api/operations/confirmations
func (h *OperationHandler) ListConfirmations(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		ParamError(c, "session_id 不能为空")
		return
	}

	limit := 50
	confirmations, err := h.confirmMgr.List(sessionID, limit)
	if err != nil {
		InternalError(c, "获取确认列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"confirmations": confirmations,
	})
}

// GetPendingConfirmations 获取待确认列表
// GET /api/operations/confirmations/pending
func (h *OperationHandler) GetPendingConfirmations(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		ParamError(c, "session_id 不能为空")
		return
	}

	confirmations, err := h.confirmMgr.GetPending(sessionID)
	if err != nil {
		InternalError(c, "获取待确认列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"confirmations": confirmations,
	})
}
