package agent

import (
	"testing"
)

func TestDoomLoopDetector_ExactMatch(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	// 连续调用相同工具和参数
	args := map[string]any{"host": "server1", "cmd": "ls"}
	detector.Record("ssh_exec", args)
	detector.Record("ssh_exec", args)
	detector.Record("ssh_exec", args)

	isLoop, reason := detector.Check()
	if !isLoop {
		t.Error("应该检测到死循环")
	}
	if reason == "" {
		t.Error("应该返回原因")
	}
}

func TestDoomLoopDetector_SameToolDifferentArgs(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	// 连续调用相同工具，不同参数
	for i := 0; i < 6; i++ {
		args := map[string]any{"host": "server1", "cmd": i}
		detector.Record("ssh_exec", args)
	}

	isLoop, reason := detector.Check()
	if !isLoop {
		t.Error("应该检测到可能的死循环")
	}
	if reason == "" {
		t.Error("应该返回原因")
	}
}

func TestDoomLoopDetector_NoLoop(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	// 调用不同工具
	detector.Record("ssh_exec", map[string]any{"cmd": "ls"})
	detector.Record("get_metrics", map[string]any{"metric": "cpu"})
	detector.Record("ssh_exec", map[string]any{"cmd": "ps"})

	isLoop, _ := detector.Check()
	if isLoop {
		t.Error("不应该检测到死循环")
	}
}

func TestDoomLoopDetector_BelowThreshold(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	// 只调用 2 次，低于阈值
	args := map[string]any{"cmd": "ls"}
	detector.Record("ssh_exec", args)
	detector.Record("ssh_exec", args)

	isLoop, _ := detector.Check()
	if isLoop {
		t.Error("未达到阈值，不应该检测到死循环")
	}
}

func TestDoomLoopDetector_Reset(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	// 记录一些调用
	args := map[string]any{"cmd": "ls"}
	detector.Record("ssh_exec", args)
	detector.Record("ssh_exec", args)
	detector.Record("ssh_exec", args)

	// 重置
	detector.Reset()

	if len(detector.history) != 0 {
		t.Error("Reset 后历史记录应该为空")
	}

	isLoop, _ := detector.Check()
	if isLoop {
		t.Error("Reset 后不应该检测到死循环")
	}
}

func TestDoomLoopDetector_WindowSize(t *testing.T) {
	detector := NewDoomLoopDetector(5, 3)

	// 记录超过窗口大小的调用
	for i := 0; i < 10; i++ {
		detector.Record("tool", map[string]any{"id": i})
	}

	if len(detector.history) > 5 {
		t.Errorf("历史记录应该限制在窗口大小内，期望 <= 5，实际 %d", len(detector.history))
	}
}

func TestDoomLoopDetector_ArgsEqual(t *testing.T) {
	detector := NewDoomLoopDetector(10, 3)

	tests := []struct {
		name     string
		a        map[string]any
		b        map[string]any
		expected bool
	}{
		{
			name:     "相同参数",
			a:        map[string]any{"host": "server1", "port": 22},
			b:        map[string]any{"host": "server1", "port": 22},
			expected: true,
		},
		{
			name:     "不同值",
			a:        map[string]any{"host": "server1"},
			b:        map[string]any{"host": "server2"},
			expected: false,
		},
		{
			name:     "不同键",
			a:        map[string]any{"host": "server1"},
			b:        map[string]any{"port": 22},
			expected: false,
		},
		{
			name:     "空 map",
			a:        map[string]any{},
			b:        map[string]any{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.argsEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("argsEqual() = %v; want %v", result, tt.expected)
			}
		})
	}
}
