package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/wordflowlab/agentsdk/pkg/types"
)

// PermissionDecision 权限决策
type PermissionDecision string

const (
	PermissionAllow PermissionDecision = "allow" // 允许
	PermissionDeny  PermissionDecision = "deny"  // 拒绝
	PermissionAsk   PermissionDecision = "ask"   // 需要询问
)

// ToolPermissionRule 工具权限规则
type ToolPermissionRule struct {
	ToolName string
	Decision PermissionDecision
	Reason   string // 规则说明
}

// ApprovalFunc 审批函数
// 返回 (decision, reason, error)
type ApprovalFunc func(ctx context.Context, call *types.ToolCallRecord) (PermissionDecision, string, error)

// PermissionHook 权限钩子
type PermissionHook struct {
	PreToolUse  PreToolUseHook  // 工具执行前
	PostToolUse PostToolUseHook // 工具执行后
}

// PreToolUseHook 工具执行前钩子
// 返回 (modified_call, error)
// 可以修改工具调用参数或返回错误来阻止执行
type PreToolUseHook func(ctx context.Context, call *types.ToolCallRecord) (*types.ToolCallRecord, error)

// PostToolUseHook 工具执行后钩子
// 可以检查结果或执行清理工作
type PostToolUseHook func(ctx context.Context, call *types.ToolCallRecord, result interface{}, err error) error

// PermissionManager 权限管理器
type PermissionManager struct {
	mu sync.RWMutex

	// 全局模式
	defaultMode types.PermissionMode

	// 工具规则
	rules map[string]*ToolPermissionRule

	// 白名单/黑名单
	allowList map[string]bool
	denyList  map[string]bool
	askList   map[string]bool

	// 审批函数
	approvalFunc ApprovalFunc

	// Hook
	hooks []PermissionHook

	// 统计
	stats PermissionStats
}

// PermissionStats 权限统计
type PermissionStats struct {
	TotalChecks    int64
	AllowedCount   int64
	DeniedCount    int64
	ApprovalCount  int64
	HookErrorCount int64
}

// PermissionManagerOptions 权限管理器配置
type PermissionManagerOptions struct {
	DefaultMode  types.PermissionMode
	AllowList    []string
	DenyList     []string
	AskList      []string
	ApprovalFunc ApprovalFunc
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager(opts *PermissionManagerOptions) *PermissionManager {
	if opts == nil {
		opts = &PermissionManagerOptions{
			DefaultMode: types.PermissionModeAuto,
		}
	}

	pm := &PermissionManager{
		defaultMode:  opts.DefaultMode,
		rules:        make(map[string]*ToolPermissionRule),
		allowList:    make(map[string]bool),
		denyList:     make(map[string]bool),
		askList:      make(map[string]bool),
		approvalFunc: opts.ApprovalFunc,
		hooks:        make([]PermissionHook, 0),
	}

	// 设置白名单
	for _, tool := range opts.AllowList {
		pm.allowList[tool] = true
	}

	// 设置黑名单
	for _, tool := range opts.DenyList {
		pm.denyList[tool] = true
	}

	// 设置审批列表
	for _, tool := range opts.AskList {
		pm.askList[tool] = true
	}

	return pm
}

// SetRule 设置工具规则
func (pm *PermissionManager) SetRule(toolName string, decision PermissionDecision, reason string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.rules[toolName] = &ToolPermissionRule{
		ToolName: toolName,
		Decision: decision,
		Reason:   reason,
	}
}

// RemoveRule 移除工具规则
func (pm *PermissionManager) RemoveRule(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.rules, toolName)
}

// AddHook 添加权限钩子
func (pm *PermissionManager) AddHook(hook PermissionHook) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.hooks = append(pm.hooks, hook)
}

// Check 检查工具权限
func (pm *PermissionManager) Check(ctx context.Context, call *types.ToolCallRecord) (PermissionDecision, string, error) {
	pm.mu.Lock()
	pm.stats.TotalChecks++
	pm.mu.Unlock()

	toolName := call.Name

	// 1. 检查黑名单 (优先级最高)
	pm.mu.RLock()
	if pm.denyList[toolName] {
		pm.mu.RUnlock()
		pm.mu.Lock()
		pm.stats.DeniedCount++
		pm.mu.Unlock()
		return PermissionDeny, "tool is in deny list", nil
	}
	pm.mu.RUnlock()

	// 2. 检查白名单
	pm.mu.RLock()
	if pm.allowList[toolName] {
		pm.mu.RUnlock()
		pm.mu.Lock()
		pm.stats.AllowedCount++
		pm.mu.Unlock()
		return PermissionAllow, "tool is in allow list", nil
	}
	pm.mu.RUnlock()

	// 3. 检查审批列表
	pm.mu.RLock()
	if pm.askList[toolName] {
		pm.mu.RUnlock()
		pm.mu.Lock()
		pm.stats.ApprovalCount++
		pm.mu.Unlock()
		return PermissionAsk, "tool requires approval", nil
	}
	pm.mu.RUnlock()

	// 4. 检查工具规则
	pm.mu.RLock()
	if rule, exists := pm.rules[toolName]; exists {
		reason := rule.Reason
		pm.mu.RUnlock()

		pm.mu.Lock()
		switch rule.Decision {
		case PermissionAllow:
			pm.stats.AllowedCount++
		case PermissionDeny:
			pm.stats.DeniedCount++
		case PermissionAsk:
			pm.stats.ApprovalCount++
		}
		pm.mu.Unlock()

		return rule.Decision, reason, nil
	}
	pm.mu.RUnlock()

	// 5. 应用全局模式
	pm.mu.RLock()
	mode := pm.defaultMode
	pm.mu.RUnlock()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch mode {
	case types.PermissionModeAllow:
		pm.stats.AllowedCount++
		return PermissionAllow, "default mode: allow", nil
	case types.PermissionModeApproval:
		pm.stats.ApprovalCount++
		return PermissionAsk, "default mode: approval", nil
	case types.PermissionModeAuto:
		// Auto 模式: 默认允许,但可以通过规则覆盖
		pm.stats.AllowedCount++
		return PermissionAllow, "default mode: auto (allow)", nil
	default:
		pm.stats.AllowedCount++
		return PermissionAllow, "default: allow", nil
	}
}

// RequestApproval 请求审批
func (pm *PermissionManager) RequestApproval(ctx context.Context, call *types.ToolCallRecord) (PermissionDecision, string, error) {
	pm.mu.RLock()
	approvalFunc := pm.approvalFunc
	pm.mu.RUnlock()

	if approvalFunc == nil {
		// 没有审批函数,默认拒绝
		return PermissionDeny, "no approval function configured", nil
	}

	decision, reason, err := approvalFunc(ctx, call)
	if err != nil {
		return PermissionDeny, fmt.Sprintf("approval error: %v", err), err
	}

	return decision, reason, nil
}

// RunPreHooks 运行前置钩子
func (pm *PermissionManager) RunPreHooks(ctx context.Context, call *types.ToolCallRecord) (*types.ToolCallRecord, error) {
	pm.mu.RLock()
	hooks := make([]PermissionHook, len(pm.hooks))
	copy(hooks, pm.hooks)
	pm.mu.RUnlock()

	modifiedCall := call
	for _, hook := range hooks {
		if hook.PreToolUse == nil {
			continue
		}

		newCall, err := hook.PreToolUse(ctx, modifiedCall)
		if err != nil {
			pm.mu.Lock()
			pm.stats.HookErrorCount++
			pm.mu.Unlock()
			return nil, fmt.Errorf("pre-hook error: %w", err)
		}

		if newCall != nil {
			modifiedCall = newCall
		}
	}

	return modifiedCall, nil
}

// RunPostHooks 运行后置钩子
func (pm *PermissionManager) RunPostHooks(ctx context.Context, call *types.ToolCallRecord, result interface{}, callErr error) error {
	pm.mu.RLock()
	hooks := make([]PermissionHook, len(pm.hooks))
	copy(hooks, pm.hooks)
	pm.mu.RUnlock()

	for _, hook := range hooks {
		if hook.PostToolUse == nil {
			continue
		}

		err := hook.PostToolUse(ctx, call, result, callErr)
		if err != nil {
			pm.mu.Lock()
			pm.stats.HookErrorCount++
			pm.mu.Unlock()
			return fmt.Errorf("post-hook error: %w", err)
		}
	}

	return nil
}

// GetStats 获取统计信息
func (pm *PermissionManager) GetStats() PermissionStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.stats
}

// SetApprovalFunc 设置审批函数
func (pm *PermissionManager) SetApprovalFunc(f ApprovalFunc) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.approvalFunc = f
}

// SetDefaultMode 设置默认模式
func (pm *PermissionManager) SetDefaultMode(mode types.PermissionMode) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.defaultMode = mode
}

// GetDefaultMode 获取默认模式
func (pm *PermissionManager) GetDefaultMode() types.PermissionMode {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.defaultMode
}

// AddToAllowList 添加到白名单
func (pm *PermissionManager) AddToAllowList(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.allowList[toolName] = true
	delete(pm.denyList, toolName)
	delete(pm.askList, toolName)
}

// AddToDenyList 添加到黑名单
func (pm *PermissionManager) AddToDenyList(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.denyList[toolName] = true
	delete(pm.allowList, toolName)
	delete(pm.askList, toolName)
}

// AddToAskList 添加到审批列表
func (pm *PermissionManager) AddToAskList(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.askList[toolName] = true
	delete(pm.allowList, toolName)
	delete(pm.denyList, toolName)
}

// RemoveFromLists 从所有列表中移除
func (pm *PermissionManager) RemoveFromLists(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.allowList, toolName)
	delete(pm.denyList, toolName)
	delete(pm.askList, toolName)
}

// IsInAllowList 检查是否在白名单
func (pm *PermissionManager) IsInAllowList(toolName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.allowList[toolName]
}

// IsInDenyList 检查是否在黑名单
func (pm *PermissionManager) IsInDenyList(toolName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.denyList[toolName]
}

// IsInAskList 检查是否在审批列表
func (pm *PermissionManager) IsInAskList(toolName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.askList[toolName]
}

// ClearStats 清空统计
func (pm *PermissionManager) ClearStats() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.stats = PermissionStats{}
}
