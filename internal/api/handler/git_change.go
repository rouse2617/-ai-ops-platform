package handler

import (
	"net/http"

	"ai-ops/internal/tool/builtin"
	"github.com/gin-gonic/gin"
)

// GitChangeHandler handles git change related requests
type GitChangeHandler struct {
	tracker *builtin.GitChangeTracker
}

// NewGitChangeHandler creates a new git change handler
func NewGitChangeHandler(tracker *builtin.GitChangeTracker) *GitChangeHandler {
	return &GitChangeHandler{
		tracker: tracker,
	}
}

// GetDiff retrieves diff for a specific commit
func (h *GitChangeHandler) GetDiff(c *gin.Context) {
	repo := c.Param("repository")
	commitHash := c.Param("commit")

	diff, err := h.tracker.GetDiff(repo, commitHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"diff": diff,
		"repository": repo,
		"commit": commitHash,
	})
}

// GetCommits retrieves commits in time range
func (h *GitChangeHandler) GetCommits(c *gin.Context) {
	var req struct {
		Repository string `json:"repository" binding:"required"`
		StartTime  string `json:"start_time" binding:"required"`
		EndTime    string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse times and call tracker
	// Implementation details omitted for brevity

	c.JSON(http.StatusOK, gin.H{
		"commits": []interface{}{},
	})
}
