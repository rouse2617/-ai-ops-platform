package handler

import (
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"

	"github.com/gin-gonic/gin"
)

// InternalSSHHandler 内部 SSH 处理器
type InternalSSHHandler struct {
	sshPool  *ssh.Pool
	hostRepo repository.HostRepository
}

// NewInternalSSHHandler 创建内部 SSH 处理器
func NewInternalSSHHandler(sshPool *ssh.Pool, hostRepo repository.HostRepository) *InternalSSHHandler {
	return &InternalSSHHandler{
		sshPool:  sshPool,
		hostRepo: hostRepo,
	}
}

// GetHostCredentials 获取主机 SSH 凭证
// GET /internal/hosts/:ip/credentials
func (h *InternalSSHHandler) GetHostCredentials(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		ParamError(c, "ip 不能为空")
		return
	}

	host, err := h.hostRepo.GetByIP(ip)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	Success(c, gin.H{
		"ip":       host.IP,
		"port":     host.Port,
		"username": host.User,
		"password": host.Password,
		"keyPath":  host.KeyPath,
		"key":      host.KeyContent,
		"authType": host.AuthType,
	})
}
