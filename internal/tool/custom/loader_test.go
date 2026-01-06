package custom

import (
	"os"
	"path/filepath"
	"testing"

	"ai-ops/internal/tool"
)

func TestCustomToolLoader_LoadFile(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(*testing.T, *CustomTool)
	}{
		{
			name: "valid tool",
			yaml: `name: test_tool
description: Test tool
parameters:
  - name: param1
    type: string
    required: true
    description: Test parameter
script: |
  #!/bin/bash
  echo "test"`,
			wantErr: false,
			check: func(t *testing.T, ct *CustomTool) {
				if ct.Name != "test_tool" {
					t.Errorf("Name = %s; want test_tool", ct.Name)
				}
				if len(ct.Parameters) != 1 {
					t.Errorf("Parameters length = %d; want 1", len(ct.Parameters))
				}
				if ct.Parameters[0].Name != "param1" {
					t.Errorf("Parameter name = %s; want param1", ct.Parameters[0].Name)
				}
			},
		},
		{
			name: "missing name",
			yaml: `description: Test
script: echo test`,
			wantErr: true,
		},
		{
			name: "missing script",
			yaml: `name: test
description: Test`,
			wantErr: true,
		},
		{
			name: "with default value",
			yaml: `name: test_tool
description: Test
parameters:
  - name: timeout
    type: int
    required: false
    description: Timeout
    default: 30
script: echo test`,
			wantErr: false,
			check: func(t *testing.T, ct *CustomTool) {
				if ct.Parameters[0].Default != 30 {
					t.Errorf("Default = %v; want 30", ct.Parameters[0].Default)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yaml), 0644); err != nil {
				t.Fatal(err)
			}

			loader := NewCustomToolLoader(tmpDir)
			ct, err := loader.LoadFile(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, ct)
			}
		})
	}
}

func TestCustomToolLoader_Load(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试工具文件
	tools := map[string]string{
		"tool1.yaml": `name: tool1
description: Tool 1
script: echo 1`,
		"tool2.yml": `name: tool2
description: Tool 2
script: echo 2`,
		"invalid.txt": "not a yaml file",
	}

	for name, content := range tools {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	loader := NewCustomToolLoader(tmpDir)
	loaded, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Load() loaded %d tools; want 2", len(loaded))
	}

	names := make(map[string]bool)
	for _, ct := range loaded {
		names[ct.Name] = true
	}

	if !names["tool1"] || !names["tool2"] {
		t.Errorf("Load() tools = %v; want tool1 and tool2", names)
	}
}

func TestCustomToolLoader_Load_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCustomToolLoader(tmpDir)
	loaded, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Load() loaded %d tools; want 0", len(loaded))
	}
}

func TestCustomToolLoader_Load_NonExistentDir(t *testing.T) {
	loader := NewCustomToolLoader("/nonexistent/path")
	loaded, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded != nil {
		t.Errorf("Load() = %v; want nil", loaded)
	}
}

func TestNewCustomToolLoader_DefaultDir(t *testing.T) {
	loader := NewCustomToolLoader("")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".ai-ops", "tools")
	if loader.toolDir != expected {
		t.Errorf("toolDir = %s; want %s", loader.toolDir, expected)
	}
}

func TestCustomTool_Parameters(t *testing.T) {
	ct := &CustomTool{
		Name:        "test",
		Description: "Test tool",
		Parameters: []tool.Parameter{
			{
				Name:        "host",
				Type:        "string",
				Required:    true,
				Description: "Target host",
			},
			{
				Name:        "port",
				Type:        "int",
				Required:    false,
				Description: "Port number",
				Default:     22,
			},
		},
		Script: "echo test",
	}

	if len(ct.Parameters) != 2 {
		t.Errorf("Parameters length = %d; want 2", len(ct.Parameters))
	}

	if ct.Parameters[0].Required != true {
		t.Error("First parameter should be required")
	}

	if ct.Parameters[1].Default != 22 {
		t.Errorf("Second parameter default = %v; want 22", ct.Parameters[1].Default)
	}
}
