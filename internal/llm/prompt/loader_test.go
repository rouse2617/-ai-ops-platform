package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptLoader_Load(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	providersDir := filepath.Join(tmpDir, "providers")
	if err := os.MkdirAll(providersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建测试模板
	baseContent := "Base prompt template"
	if err := os.WriteFile(filepath.Join(tmpDir, "base.txt"), []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	openaiContent := "OpenAI specific prompt"
	if err := os.WriteFile(filepath.Join(providersDir, "openai.txt"), []byte(openaiContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewPromptLoader(tmpDir)

	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{"load openai", "openai", openaiContent},
		{"fallback to base", "unknown", baseContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.Load(tt.provider)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptLoader_Render(t *testing.T) {
	tmpDir := t.TempDir()
	providersDir := filepath.Join(tmpDir, "providers")
	if err := os.MkdirAll(providersDir, 0755); err != nil {
		t.Fatal(err)
	}

	tmplContent := "Tools: {{.Tools}}\nContext: {{.Context}}"
	if err := os.WriteFile(filepath.Join(providersDir, "test.txt"), []byte(tmplContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewPromptLoader(tmpDir)

	data := map[string]any{
		"Tools":   "tool1, tool2",
		"Context": "test context",
	}

	got, err := loader.Render("test", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "Tools: tool1, tool2\nContext: test context"
	if got != want {
		t.Errorf("Render() = %v, want %v", got, want)
	}
}

func TestPromptLoader_Cache(t *testing.T) {
	tmpDir := t.TempDir()
	providersDir := filepath.Join(tmpDir, "providers")
	if err := os.MkdirAll(providersDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "Cached template"
	if err := os.WriteFile(filepath.Join(providersDir, "cached.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewPromptLoader(tmpDir)

	// 第一次加载
	_, err := loader.Load("cached")
	if err != nil {
		t.Fatalf("First Load() error = %v", err)
	}

	// 验证缓存
	if _, ok := loader.cache["cached"]; !ok {
		t.Error("Template not cached")
	}

	// 第二次加载应该使用缓存
	_, err = loader.Load("cached")
	if err != nil {
		t.Fatalf("Second Load() error = %v", err)
	}
}

func TestPromptLoader_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	providersDir := filepath.Join(tmpDir, "providers")
	if err := os.MkdirAll(providersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建无效模板
	invalidContent := "{{.Invalid"
	if err := os.WriteFile(filepath.Join(providersDir, "invalid.txt"), []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewPromptLoader(tmpDir)

	_, err := loader.Load("invalid")
	if err == nil {
		t.Error("Expected error for invalid template, got nil")
	}
}

func TestPromptLoader_MissingBaseTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewPromptLoader(tmpDir)

	_, err := loader.Load("nonexistent")
	if err == nil {
		t.Error("Expected error for missing base template, got nil")
	}
}
