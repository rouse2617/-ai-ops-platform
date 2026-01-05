package builtin

import (
	"testing"
)

func TestCheckMySQLStatusTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "缺少 host 参数",
			params: map[string]interface{}{
				"mysql_user": "root",
			},
			expectError: true,
			errorMsg:    "请指定 host 参数",
		},
		{
			name: "缺少 mysql_user 参数",
			params: map[string]interface{}{
				"host": "test-host",
			},
			expectError: true,
			errorMsg:    "请指定 mysql_user 参数",
		},
		{
			name: "参数完整但无 SSH 上下文",
			params: map[string]interface{}{
				"host":       "test-host",
				"mysql_user": "root",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	mysqlTool := NewCheckMySQLStatusTool()

	if mysqlTool.Name() != "check_mysql_status" {
		t.Errorf("Name() = %s; want check_mysql_status", mysqlTool.Name())
	}
	if mysqlTool.Description() == "" {
		t.Error("Description() should not be empty")
	}
	params := mysqlTool.Parameters()
	if len(params) != 5 {
		t.Errorf("Parameters() length = %d; want 5", len(params))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mysqlTool.Execute(nil, tt.params)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if tt.expectError {
				if result.Success {
					t.Error("Expected error result, got success")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("Error message = %s; want %s", result.Error, tt.errorMsg)
				}
			}
		})
	}
}

func TestCheckRedisStatusTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "缺少 host 参数",
			params:      map[string]interface{}{},
			expectError: true,
			errorMsg:    "请指定 host 参数",
		},
		{
			name: "参数完整但无 SSH 上下文",
			params: map[string]interface{}{
				"host": "test-host",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	redisTool := NewCheckRedisStatusTool()

	if redisTool.Name() != "check_redis_status" {
		t.Errorf("Name() = %s; want check_redis_status", redisTool.Name())
	}
	if redisTool.Description() == "" {
		t.Error("Description() should not be empty")
	}
	params := redisTool.Parameters()
	if len(params) != 4 {
		t.Errorf("Parameters() length = %d; want 4", len(params))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := redisTool.Execute(nil, tt.params)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if tt.expectError {
				if result.Success {
					t.Error("Expected error result, got success")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("Error message = %s; want %s", result.Error, tt.errorMsg)
				}
			}
		})
	}
}

func TestBackupDatabaseTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "缺少所有参数",
			params:      map[string]interface{}{},
			expectError: true,
			errorMsg:    "请指定 host 参数",
		},
		{
			name: "缺少 db_type 参数",
			params: map[string]interface{}{
				"host": "test-host",
			},
			expectError: true,
			errorMsg:    "请指定 db_type 参数（mysql 或 postgresql）",
		},
		{
			name: "参数完整但无 SSH 上下文",
			params: map[string]interface{}{
				"host":        "test-host",
				"db_type":     "mysql",
				"db_name":     "testdb",
				"db_user":     "root",
				"db_password": "password",
				"backup_path": "/tmp/backup.sql",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	backupTool := NewBackupDatabaseTool()

	if backupTool.Name() != "backup_database" {
		t.Errorf("Name() = %s; want backup_database", backupTool.Name())
	}
	if backupTool.Description() == "" {
		t.Error("Description() should not be empty")
	}
	params := backupTool.Parameters()
	if len(params) != 6 {
		t.Errorf("Parameters() length = %d; want 6", len(params))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := backupTool.Execute(nil, tt.params)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if tt.expectError {
				if result.Success {
					t.Error("Expected error result, got success")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("Error message = %s; want %s", result.Error, tt.errorMsg)
				}
			}
		})
	}
}
