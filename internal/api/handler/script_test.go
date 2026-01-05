package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"ai-ops/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	CodeSuccess    = int(errors.CodeSuccess)
	CodeParamError = int(errors.CodeParamError)
	CodeNotFound   = int(errors.CodeNotFound)
)

func setupScriptHandler() (*ScriptHandler, *gin.Engine) {
	scriptRepo := newMockScriptRepository()
	hostRepo := newMockHostRepository()
	registry := tool.NewRegistry()
	pool := ssh.NewPool(ssh.Config{})
	handler := NewScriptHandler(scriptRepo, hostRepo, registry, pool)

	r := gin.New()
	api := r.Group("/api")
	{
		scripts := api.Group("/scripts")
		{
			scripts.GET("", handler.ListScripts)
			scripts.POST("", handler.CreateScript)
			scripts.GET("/:id", handler.GetScript)
			scripts.PUT("/:id", handler.UpdateScript)
			scripts.DELETE("/:id", handler.DeleteScript)
			scripts.PUT("/:id/toggle", handler.ToggleScript)
			scripts.POST("/upload", handler.UploadScript)
			scripts.POST("/:id/test", handler.TestScript)
		}
	}

	return handler, r
}

// TestListScripts_Empty 测试空脚本列表
func TestListScripts_Empty(t *testing.T) {
	_, r := setupScriptHandler()

	req, _ := http.NewRequest("GET", "/api/scripts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.NotNil(t, data["list"])
	assert.Equal(t, float64(0), data["total"])
}

// TestCreateScript_Success 测试创建脚本成功
func TestCreateScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	body := map[string]interface{}{
		"name":        "test-script",
		"description": "测试脚本",
		"language":    "bash",
		"content":     "echo 'hello'",
		"enabled":     true,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "test-script", data["name"])
	assert.Equal(t, "bash", data["language"])
	assert.Equal(t, "echo 'hello'", data["content"])
}

// TestCreateScript_MissingName 测试创建脚本缺少名称
func TestCreateScript_MissingName(t *testing.T) {
	_, r := setupScriptHandler()

	body := map[string]interface{}{
		"language": "bash",
		"content":  "echo 'hello'",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeParamError, resp.Code)
}

// TestCreateScript_InvalidLanguage 测试创建脚本语言无效
func TestCreateScript_InvalidLanguage(t *testing.T) {
	_, r := setupScriptHandler()

	body := map[string]interface{}{
		"name":     "test-script",
		"language": "invalid",
		"content":  "echo 'hello'",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeParamError, resp.Code)
}

// TestGetScript_Success 测试获取单个脚本成功
func TestGetScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	// 先创建一个脚本
	body := map[string]interface{}{
		"name":     "get-test-script",
		"language": "bash",
		"content":  "echo 'test'",
		"enabled":  true,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	scriptData := createResp.Data.(map[string]interface{})
	scriptID := scriptData["id"].(string)

	// 获取脚本
	req, _ = http.NewRequest("GET", "/api/scripts/"+scriptID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "get-test-script", data["name"])
}

// TestGetScript_NotFound 测试获取不存在的脚本
func TestGetScript_NotFound(t *testing.T) {
	_, r := setupScriptHandler()

	req, _ := http.NewRequest("GET", "/api/scripts/non-existent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeNotFound, resp.Code)
}

// TestUpdateScript_Success 测试更新脚本成功
func TestUpdateScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	// 先创建一个脚本
	body := map[string]interface{}{
		"name":     "update-test-script",
		"language": "bash",
		"content":  "echo 'old'",
		"enabled":  true,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	scriptData := createResp.Data.(map[string]interface{})
	scriptID := scriptData["id"].(string)

	// 更新脚本
	updateBody := map[string]interface{}{
		"name":     "update-test-script",
		"content":  "echo 'new'",
		"enabled":  false,
		"language": "bash",
	}
	jsonBody, _ = json.Marshal(updateBody)
	req, _ = http.NewRequest("PUT", "/api/scripts/"+scriptID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "echo 'new'", data["content"])
}

// TestDeleteScript_Success 测试删除脚本成功
func TestDeleteScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	// 先创建一个脚本
	body := map[string]interface{}{
		"name":     "delete-test-script",
		"language": "bash",
		"content":  "echo 'test'",
		"enabled":  true,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	scriptData := createResp.Data.(map[string]interface{})
	scriptID := scriptData["id"].(string)

	// 删除脚本
	req, _ = http.NewRequest("DELETE", "/api/scripts/"+scriptID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	// 验证已删除
	req, _ = http.NewRequest("GET", "/api/scripts/"+scriptID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeNotFound, resp.Code)
}

// TestToggleScript_Success 测试启用/禁用脚本成功
func TestToggleScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	// 先创建一个脚本
	body := map[string]interface{}{
		"name":     "toggle-test-script",
		"language": "bash",
		"content":  "echo 'test'",
		"enabled":  true,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	scriptData := createResp.Data.(map[string]interface{})
	scriptID := scriptData["id"].(string)

	// 禁用脚本
	toggleBody := map[string]interface{}{
		"enabled": false,
	}
	jsonBody, _ = json.Marshal(toggleBody)
	req, _ = http.NewRequest("PUT", "/api/scripts/"+scriptID+"/toggle", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, false, data["enabled"])
}

// TestListScripts_WithKeyword 测试关键字过滤
func TestListScripts_WithKeyword(t *testing.T) {
	_, r := setupScriptHandler()

	// 创建多个脚本
	scripts := []map[string]interface{}{
		{"name": "backup-script", "language": "bash", "content": "backup", "enabled": true},
		{"name": "deploy-script", "language": "bash", "content": "deploy", "enabled": true},
		{"name": "backup-db", "language": "python", "content": "backup db", "enabled": true},
	}

	for _, s := range scripts {
		jsonBody, _ := json.Marshal(s)
		req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 搜索 backup
	req, _ := http.NewRequest("GET", "/api/scripts?keyword=backup", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(2), data["total"])
}

// TestUploadScript_Success 测试上传脚本成功
func TestUploadScript_Success(t *testing.T) {
	_, r := setupScriptHandler()

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.sh")
	part.Write([]byte("#!/bin/bash\necho 'test'"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/scripts/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "test", data["name"])
	assert.Equal(t, "bash", data["language"])
}

// TestUploadScript_InvalidExtension 测试上传不支持的文件类型
func TestUploadScript_InvalidExtension(t *testing.T) {
	_, r := setupScriptHandler()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/scripts/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeParamError, resp.Code)
}

// TestListScripts_AfterCreate 测试创建后列表
func TestListScripts_AfterCreate(t *testing.T) {
	_, r := setupScriptHandler()

	// 创建脚本
	body := map[string]interface{}{
		"name":     "list-test-script",
		"language": "bash",
		"content":  "echo 'test'",
		"enabled":  true,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/scripts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 获取列表
	req, _ = http.NewRequest("GET", "/api/scripts", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, CodeSuccess, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}
