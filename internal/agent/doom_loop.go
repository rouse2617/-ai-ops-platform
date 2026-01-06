package agent

import (
	"encoding/json"
	"fmt"
)

// ToolCall 工具调用记录
type ToolCall struct {
	Name string
	Args map[string]any
}

// DoomLoopDetector 死循环检测器
type DoomLoopDetector struct {
	windowSize int        // 检测窗口大小
	threshold  int        // 重复次数阈值
	history    []ToolCall // 调用历史
}

// NewDoomLoopDetector 创建检测器
func NewDoomLoopDetector(windowSize, threshold int) *DoomLoopDetector {
	return &DoomLoopDetector{
		windowSize: windowSize,
		threshold:  threshold,
		history:    make([]ToolCall, 0, windowSize),
	}
}

// Record 记录工具调用
func (d *DoomLoopDetector) Record(toolName string, args map[string]any) {
	d.history = append(d.history, ToolCall{Name: toolName, Args: args})
	if len(d.history) > d.windowSize {
		d.history = d.history[1:]
	}
}

// Check 检测是否陷入死循环
func (d *DoomLoopDetector) Check() (isLoop bool, reason string) {
	if len(d.history) < d.threshold {
		return false, ""
	}

	// 检测相同工具+相同参数
	exactCount := d.countExactMatches()
	if exactCount >= d.threshold {
		return true, fmt.Sprintf("连续 %d 次调用相同工具和参数: %s", exactCount, d.history[len(d.history)-1].Name)
	}

	// 检测相同工具（不同参数）
	sameToolCount := d.countSameToolCalls()
	if sameToolCount >= d.threshold*2 {
		return true, fmt.Sprintf("连续 %d 次调用相同工具: %s (可能参数不同)", sameToolCount, d.history[len(d.history)-1].Name)
	}

	return false, ""
}

// Reset 重置检测器
func (d *DoomLoopDetector) Reset() {
	d.history = make([]ToolCall, 0, d.windowSize)
}

// countExactMatches 统计完全相同的连续调用
func (d *DoomLoopDetector) countExactMatches() int {
	if len(d.history) == 0 {
		return 0
	}

	last := d.history[len(d.history)-1]
	count := 1

	for i := len(d.history) - 2; i >= 0; i-- {
		if d.history[i].Name == last.Name && d.argsEqual(d.history[i].Args, last.Args) {
			count++
		} else {
			break
		}
	}

	return count
}

// countSameToolCalls 统计相同工具的连续调用
func (d *DoomLoopDetector) countSameToolCalls() int {
	if len(d.history) == 0 {
		return 0
	}

	lastTool := d.history[len(d.history)-1].Name
	count := 1

	for i := len(d.history) - 2; i >= 0; i-- {
		if d.history[i].Name == lastTool {
			count++
		} else {
			break
		}
	}

	return count
}

// argsEqual 比较参数是否相等
func (d *DoomLoopDetector) argsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	// 简单序列化比较
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}
