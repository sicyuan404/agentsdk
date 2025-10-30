package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wordflowlab/agentsdk/pkg/types"
)

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	MaxConcurrency int           // 最大并发数
	DefaultTimeout time.Duration // 默认超时时间
}

// Executor 工具执行器
type Executor struct {
	config   ExecutorConfig
	semaphore chan struct{}
	running   sync.WaitGroup
}

// NewExecutor 创建工具执行器
func NewExecutor(config ExecutorConfig) *Executor {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 3 // 默认3个并发
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 60 * time.Second
	}

	return &Executor{
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrency),
	}
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
	Tool    Tool
	Input   map[string]interface{}
	Context *ToolContext
	Timeout time.Duration
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	Success    bool
	Output     interface{}
	Error      error
	DurationMs int64
	StartedAt  time.Time
	EndedAt    time.Time
}

// Execute 执行单个工具
func (e *Executor) Execute(ctx context.Context, req *ExecuteRequest) *ExecuteResult {
	startTime := time.Now()

	// 获取信号量
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return &ExecuteResult{
			Success:   false,
			Error:     ctx.Err(),
			StartedAt: startTime,
			EndedAt:   time.Now(),
		}
	}

	// 设置超时
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = e.config.DefaultTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行工具
	output, err := req.Tool.Execute(execCtx, req.Input, req.Context)
	endTime := time.Now()

	result := &ExecuteResult{
		Success:    err == nil,
		Output:     output,
		Error:      err,
		StartedAt:  startTime,
		EndedAt:    endTime,
		DurationMs: endTime.Sub(startTime).Milliseconds(),
	}

	return result
}

// ExecuteBatch 批量执行工具
func (e *Executor) ExecuteBatch(ctx context.Context, requests []*ExecuteRequest) []*ExecuteResult {
	results := make([]*ExecuteResult, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, r *ExecuteRequest) {
			defer wg.Done()
			results[idx] = e.Execute(ctx, r)
		}(i, req)
	}

	wg.Wait()
	return results
}

// Wait 等待所有执行完成
func (e *Executor) Wait() {
	e.running.Wait()
}

// ValidateInput 验证工具输入
func ValidateInput(tool Tool, input map[string]interface{}) error {
	schema := tool.InputSchema()
	if schema == nil {
		return nil // 没有schema,跳过验证
	}

	// TODO: 使用jsonschema库进行验证
	// 这里先做简单的required字段检查
	if required, ok := schema["required"].([]interface{}); ok {
		for _, field := range required {
			fieldName := field.(string)
			if _, exists := input[fieldName]; !exists {
				return fmt.Errorf("missing required field: %s", fieldName)
			}
		}
	}

	return nil
}

// ToolCallRecordBuilder 工具调用记录构建器
type ToolCallRecordBuilder struct {
	record *types.ToolCallRecord
}

// NewToolCallRecord 创建工具调用记录
func NewToolCallRecord(id, name string, input map[string]interface{}) *ToolCallRecordBuilder {
	now := time.Now()
	return &ToolCallRecordBuilder{
		record: &types.ToolCallRecord{
			ID:        id,
			Name:      name,
			Input:     input,
			State:     types.ToolCallStatePending,
			Approval:  types.ToolCallApproval{Required: false},
			CreatedAt: now,
			UpdatedAt: now,
			AuditTrail: []types.ToolCallAuditEntry{
				{
					State:     types.ToolCallStatePending,
					Timestamp: now,
					Note:      "created",
				},
			},
		},
	}
}

// SetState 设置状态
func (b *ToolCallRecordBuilder) SetState(state types.ToolCallState, note string) *ToolCallRecordBuilder {
	now := time.Now()
	b.record.State = state
	b.record.UpdatedAt = now
	b.record.AuditTrail = append(b.record.AuditTrail, types.ToolCallAuditEntry{
		State:     state,
		Timestamp: now,
		Note:      note,
	})
	return b
}

// SetApproval 设置审批信息
func (b *ToolCallRecordBuilder) SetApproval(approval types.ToolCallApproval) *ToolCallRecordBuilder {
	b.record.Approval = approval
	return b
}

// SetResult 设置结果
func (b *ToolCallRecordBuilder) SetResult(result interface{}, err error) *ToolCallRecordBuilder {
	if err != nil {
		b.record.Error = err.Error()
		b.record.IsError = true
		b.SetState(types.ToolCallStateFailed, "execution failed")
	} else {
		b.record.Result = result
		b.SetState(types.ToolCallStateCompleted, "execution succeeded")
	}
	return b
}

// SetTiming 设置时间信息
func (b *ToolCallRecordBuilder) SetTiming(startedAt, completedAt time.Time) *ToolCallRecordBuilder {
	b.record.StartedAt = &startedAt
	b.record.CompletedAt = &completedAt
	duration := completedAt.Sub(startedAt).Milliseconds()
	b.record.DurationMs = &duration
	return b
}

// Build 构建记录
func (b *ToolCallRecordBuilder) Build() *types.ToolCallRecord {
	return b.record
}
