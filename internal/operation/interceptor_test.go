package operation

import (
	"context"
	"testing"
	"time"

	"ai-ops/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestDangerInterceptor_CheckCommand(t *testing.T) {
	interceptor := NewDangerInterceptor(nil)

	tests := []struct {
		name          string
		command       string
		isDangerous   bool
		expectedLevel string
	}{
		{
			name:          "删除根目录",
			command:       "rm -rf /",
			isDangerous:   true,
			expectedLevel: model.RiskLevelCritical,
		},
		{
			name:          "删除系统目录",
			command:       "rm -rf /etc",
			isDangerous:   true,
			expectedLevel: model.RiskLevelCritical,
		},
		{
			name:          "格式化磁盘",
			command:       "mkfs.ext4 /dev/sda1",
			isDangerous:   true,
			expectedLevel: model.RiskLevelCritical,
		},
		{
			name:          "关机",
			command:       "shutdown -h now",
			isDangerous:   true,
			expectedLevel: model.RiskLevelHigh,
		},
		{
			name:          "重启",
			command:       "reboot",
			isDangerous:   true,
			expectedLevel: model.RiskLevelHigh,
		},
		{
			name:          "安全命令 - ls",
			command:       "ls -la",
			isDangerous:   false,
			expectedLevel: "",
		},
		{
			name:          "安全命令 - ps",
			command:       "ps aux",
			isDangerous:   false,
			expectedLevel: "",
		},
		{
			name:          "安全命令 - df",
			command:       "df -h",
			isDangerous:   false,
			expectedLevel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDangerous, riskLevel, _ := interceptor.CheckCommand(tt.command)
			assert.Equal(t, tt.isDangerous, isDangerous, "命令危险性判断错误")
			if tt.isDangerous {
				assert.Equal(t, tt.expectedLevel, riskLevel, "风险级别判断错误")
			}
		})
	}
}

func TestDangerInterceptor_AddPattern(t *testing.T) {
	interceptor := NewDangerInterceptor(nil)

	err := interceptor.AddPattern(`\bkill\s+-9\s+1$`, model.RiskLevelCritical, "终止 init 进程")
	assert.NoError(t, err)

	isDangerous, riskLevel, reason := interceptor.CheckCommand("kill -9 1")
	assert.True(t, isDangerous)
	assert.Equal(t, model.RiskLevelCritical, riskLevel)
	assert.Equal(t, "终止 init 进程", reason)
}

func TestDangerInterceptor_RemovePattern(t *testing.T) {
	interceptor := NewDangerInterceptor(nil)

	pattern := `\btest_pattern\b`
	err := interceptor.AddPattern(pattern, model.RiskLevelMedium, "测试模式")
	assert.NoError(t, err)

	isDangerous, _, _ := interceptor.CheckCommand("test_pattern")
	assert.True(t, isDangerous)

	interceptor.RemovePattern(pattern)

	isDangerous, _, _ = interceptor.CheckCommand("test_pattern")
	assert.False(t, isDangerous)
}

type mockNotifier struct {
	called bool
}

func (m *mockNotifier) NotifyConfirmationRequest(confirmation *model.OperationConfirmation) error {
	m.called = true
	return nil
}

func TestConfirmationManager_RequestConfirmation(t *testing.T) {
	// 这个测试需要数据库，这里只做基本的逻辑测试
	t.Skip("需要数据库支持")
}

func TestConfirmationManager_Approve(t *testing.T) {
	t.Skip("需要数据库支持")
}

func TestConfirmationManager_Reject(t *testing.T) {
	t.Skip("需要数据库支持")
}

func TestConfirmationManager_Timeout(t *testing.T) {
	t.Skip("需要数据库支持")
}

// 基准测试
func BenchmarkDangerInterceptor_CheckCommand(b *testing.B) {
	interceptor := NewDangerInterceptor(nil)
	command := "rm -rf /tmp/test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interceptor.CheckCommand(command)
	}
}

func BenchmarkDangerInterceptor_CheckCommand_Safe(b *testing.B) {
	interceptor := NewDangerInterceptor(nil)
	command := "ls -la /tmp"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interceptor.CheckCommand(command)
	}
}

// 并发测试
func TestDangerInterceptor_Concurrent(t *testing.T) {
	interceptor := NewDangerInterceptor(nil)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			interceptor.CheckCommand("rm -rf /")
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestDangerInterceptor_AddPattern_Concurrent(t *testing.T) {
	interceptor := NewDangerInterceptor(nil)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			err := interceptor.AddPattern(`\btest`+string(rune(idx))+`\b`, model.RiskLevelMedium, "测试")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	patterns := interceptor.ListPatterns()
	assert.GreaterOrEqual(t, len(patterns), 10)
}

// 集成测试示例
func TestDangerInterceptor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建完整的拦截器和确认管理器
	notifier := &mockNotifier{}
	// confirmMgr := NewConfirmationManager(repo, notifier)
	// interceptor := NewDangerInterceptor(confirmMgr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试拦截流程
	_ = ctx
	_ = notifier
}
